#!/bin/bash

update_repository(){
  # This is the copy of the tree used to run gpackages against.
  if [[ ! -d /mnt/packages-tree/gentoo/ ]]; then
      cd /mnt/packages-tree || exit 1
      git clone https://anongit.gentoo.org/git/repo/gentoo.git
  else
      cd /mnt/packages-tree/gentoo/ || exit 1
      git pull origin master --rebase &>/dev/null
  fi
}

update_md5cache(){
  mkdir -p /var/cache/pgo-egencache
  cd /mnt/packages-tree/gentoo/ || exit 1

  #echo 'FEATURES="-userpriv -usersandbox -sandbox"' >> /etc/portage/make.conf

  egencache -j 6 --cache-dir /var/cache/pgo-egencache --repo gentoo --repositories-configuration '[gentoo]
  location = /mnt/packages-tree/gentoo' --update

  egencache -j 6 --cache-dir /var/cache/pgo-egencache --repo gentoo --repositories-configuration '[gentoo]
  location = /mnt/packages-tree/gentoo' --update-use-local-desc
}

fullupdate_database(){
  cd /mnt/packages-tree/gentoo/ || exit 1
  /go/src/soko/bin/soko --fullupdate
}


update_repository
update_md5cache
fullupdate_database
