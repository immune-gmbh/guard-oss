#!/bin/sh

remove() {
  for i in `dkms status | grep flash_mmap | cut -f 1 -d,`
  do
    dkms remove $i --all
  done
}

action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  action="remove"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
  # deb passes $1=configure $2=<current version>
  action="upgrade"
fi

case "$action" in
  "0" | "remove")
    printf "\033[32mRemoving flash_mmap DKMS module\033[0m\n"
    remove
    ;;
  "1" | "upgrade")
    remove
    ;;
  *)
    printf "\033[32mRemoving flash_mmap DKMS module\033[0m\n"
    remove
    ;;
esac