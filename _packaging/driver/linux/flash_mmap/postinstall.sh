#!/bin/sh

install() {
    dkms ldtarball /usr/src/flash_mmap-1.0.dkms.tar.gz
    dkms autoinstall
}

action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
  # deb passes $1=configure $2=<current version>
  action="upgrade"
fi

case "$action" in
  "1" | "install")
    printf "\033[32mInstalling flash_mmap via DKMS\033[0m\n"
    install
    ;;
  "2" | "upgrade")
    printf "\033[32mInstalling flash_mmap via DKMS\033[0m\n"
    install
    ;;
  *)
    printf "\033[32mInstalling flash_mmap via DKMS\033[0m\n"
    install
    ;;
esac