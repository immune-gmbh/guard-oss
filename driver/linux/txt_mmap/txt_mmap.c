#include <linux/version.h>
#include <linux/init.h>
#include <linux/module.h>
#include <asm/io.h>
#include <linux/sysfs.h> 
#include <linux/kobject.h>

#define TXT_REGION_START 0xFED30000ULL
#define TXT_REGION_SIZE 0x10000ULL
#define TXT_REGION_MASK (TXT_REGION_SIZE-1)

struct kobject *kobj_ref;

static ssize_t public_space_read(struct file *file, struct kobject *kobj,
				struct bin_attribute *bin_attr, char *buffer,
				loff_t offset, size_t count)
{
	resource_size_t pa;
	size_t copysize, remapsize;
	void __iomem *va;

	offset = offset & TXT_REGION_MASK;
	pa = (TXT_REGION_START + offset) & PAGE_MASK;

	if((offset + count) > TXT_REGION_SIZE)
		copysize = TXT_REGION_SIZE - offset;
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

BIN_ATTR_RO(public_space, TXT_REGION_SIZE);

static int __init txt_mmap_init(void)
{
	int ret = 0;

	/* create sysfs entry under kernel object */
	kobj_ref = kobject_create_and_add("txt_mmap", kernel_kobj);
	if(!kobj_ref)
	{
		pr_err("kobject_create_and_add failed\n");
		goto out;
	}

	/* extend with a single binary attribute to read the flash */
	ret = sysfs_create_bin_file(kobj_ref, &bin_attr_public_space);
	if(ret)
	{
			pr_err("sysfs_create_bin_file failed\n");
			goto error;
	}
	
	return ret;

error:
	kobject_put(kobj_ref);
out:
	return ret;
}

static void __exit txt_mmap_exit(void)
{
	if(kobj_ref)
	{
        sysfs_remove_bin_file(kernel_kobj, &bin_attr_public_space);
        kobject_put(kobj_ref); 
	}
}

module_init(txt_mmap_init);
module_exit(txt_mmap_exit);
MODULE_DESCRIPTION("txt public space mmap driver");
MODULE_AUTHOR("Hans-Gert Dahmen <hans-gert.dahmen@immu.ne>");
MODULE_LICENSE("GPL");

