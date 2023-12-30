// SPDX-License-Identifier: GPL-2.0
/*
 * Export the memory-mapped BIOS region of the platform SPI flash as
 * a read-only sysfs binary attribute on X86_64 systems.
 *
 * Copyright © 2021 immune GmbH
 */

#include <linux/version.h>
#include <linux/init.h>
#include <linux/module.h>
#include <linux/io.h>
#include <linux/sysfs.h>
#include <linux/kobject.h>

#define FLASH_REGION_START 0xFF000000ULL
#define FLASH_REGION_SIZE 0x1000000ULL
#define FLASH_REGION_MASK (FLASH_REGION_SIZE - 1)

struct kobject *kobj_ref;

static ssize_t bios_region_read(struct file *file, struct kobject *kobj,
				struct bin_attribute *bin_attr, char *buffer,
				loff_t offset, size_t count)
{
	resource_size_t pa;
	size_t copysize, remapsize;
	void __iomem *va;

	offset = offset & FLASH_REGION_MASK;
	pa = (FLASH_REGION_START + offset) & PAGE_MASK;

	if((offset + count) > FLASH_REGION_SIZE)
		copysize = FLASH_REGION_SIZE - offset;
	else
		copysize = min(count, PAGE_SIZE);

	if(((offset & ~PAGE_MASK) + copysize) > PAGE_SIZE)
		remapsize = 2 * PAGE_SIZE;
	else
		remapsize = PAGE_SIZE;

	va = ioremap(pa, remapsize);
	memcpy_fromio(buffer, va, copysize);
	iounmap(va);

	return copysize;
}

BIN_ATTR_RO(bios_region, FLASH_REGION_SIZE);

static int __init flash_mmap_init(void)
{
	int ret = 0;

	/* create sysfs entry under kernel object */
	kobj_ref = kobject_create_and_add("flash_mmap", firmware_kobj);
	if(!kobj_ref) {
		pr_err("kobject_create_and_add failed\n");
		ret = EINVAL;
		goto out;
	}

	/* extend with a single binary attribute to read the flash */
	ret = sysfs_create_bin_file(kobj_ref, &bin_attr_bios_region);
	if(ret) {
		pr_err("sysfs_create_bin_file failed\n");
		goto error;
	}

	return ret;
	
error:
	kobject_put(kobj_ref);
out:
	return ret;
}

static void __exit flash_mmap_exit(void)
{
	if(kobj_ref)
	{
		sysfs_remove_bin_file(kobj_ref, &bin_attr_bios_region);
		kobject_put(kobj_ref);
	}
}

module_init(flash_mmap_init);
module_exit(flash_mmap_exit);
MODULE_DESCRIPTION("Export SPI platform flash memory mapped region via sysfs");
MODULE_AUTHOR("Hans-Gert Dahmen <hans-gert.dahmen@immu.ne>");
MODULE_LICENSE("GPL");
