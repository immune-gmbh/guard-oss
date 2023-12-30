RELEASE_ID:=1.0
MODNAME:=flash_mmap
MOD:=$(MODNAME)-$(RELEASE_ID)
MODSRC:=../../../../driver/linux/$(MOD)

# command names
GO:=go
NFPM:=nfpm

# build-time parameters, values
export RELEASE_ID := $(RELEASE_ID)
NFPMFLAGS=package -f nfpm-packages.yml

# build artifacts
SRCS:=$(MOD).dkms.tar.gz
TARGETS:=flash_mmap_module-1-$(RELEASE_ID)_amd64.deb flash_mmap_module-1-$(RELEASE_ID).x86_64.rpm

.PHONY: deps
deps:
	@echo "Fetching and installing dependencies.."
	@$(GO) install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.19.2

# suppress lots of legacy SCCS and RCS lookups
MAKEFLAGS += --no-builtin-rules

.DEFAULT_GOAL:=all
.PHONY: all
all: $(TARGETS)

dkms:
	mkdir -p dkms dkms_src
	make -C $(MODSRC) clean
	fakeroot dkms --sourcetree $(PWD)/dkms_src --dkmstree $(PWD)/dkms add $(MODSRC)

$(MOD).dkms.tar.gz: |dkms
	dkms --sourcetree $(PWD)/dkms_src --dkmstree $(PWD)/dkms --source-only --archive $(PWD)/$(MOD).dkms.tar.gz mktarball $(MODNAME)/$(RELEASE_ID)

%.deb: $(MOD).dkms.tar.gz deps
	$(NFPM) $(NFPMFLAGS) -p deb -t $@

%.rpm: $(MOD).dkms.tar.gz deps
	$(NFPM) $(NFPMFLAGS) -p rpm -t $@

.PHONY: clean
clean:
	rm -f $(TARGETS)
	rm -rf dkms dkms_src $(MOD).dkms.tar.gz

# disable many builtin rules
.SUFFIXES:
