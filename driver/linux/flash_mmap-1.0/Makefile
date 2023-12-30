MOD = flash_mmap
KDIR ?= /lib/modules/$(shell uname -r)/build
PWD :=$(shell pwd)

all: module

module: 
	make -C $(KDIR) M=$(PWD) modules

clean:
	make -C $(KDIR) O=$(PWD)/build M=$(PWD) clean
