#!/bin/bash

update_repository(){
  # This is the copy of the tree used to run gpackages against.
  if [[ ! -d /mnt/packages-tree/gentoo/ ]]; then
      cd /mnt/packages-tree || exit 1
      git clone https://anongit.gentoo.org/git/repo/gentoo.git
  else
      cd /mnt/packages-tree/gentoo/ || exit 1
      git pull --rebase &>/dev/null
  fi
}

cleanup_database(){
  cd /mnt/packages-tree/gentoo/ || exit 1
  /go/src/soko/bin/soko fullupdate
}


update_repository
cleanup_database
