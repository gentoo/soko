#!/bin/bash

: "${GIT_URI:=https://anongit.gentoo.org/git/repo/gentoo.git}"
: "${GIT_BRANCH:=master}"
: "${GIT_REMOTE:=origin}"
: "${JOBS:=6}"

update_repository(){
  # This is the copy of the tree used to run gpackages against.
  if [[ ! -d /mnt/packages-tree/gentoo/ ]]; then
      cd /mnt/packages-tree || exit 1
      git clone \
        --quiet \
        --single-branch \
        --branch "${GIT_BRANCH}" \
        --origin "${GIT_REMOTE}" \
        "${GIT_URI}"
  else
      cd /mnt/packages-tree/gentoo/ || exit 1
      if [ "$(git remote get-url "${GIT_REMOTE}")" != "${GIT_URI}" ]; then
          git remote set-url "${GIT_REMOTE}" "${GIT_URI}"
      fi
      git fetch --quiet --force "${GIT_REMOTE}" "${GIT_BRANCH}"
      git reset --quiet --hard "${GIT_REMOTE}"/"${GIT_BRANCH}"
  fi
}

update_md5cache(){
  mkdir -p /var/cache/pgo-egencache
  cd /mnt/packages-tree/gentoo/ || exit 1

  #echo 'FEATURES="-userpriv -usersandbox -sandbox"' >> /etc/portage/make.conf

  egencache -j "${JOBS}" --cache-dir /var/cache/pgo-egencache --repo gentoo --repositories-configuration '[gentoo]
  location = /mnt/packages-tree/gentoo' --update

  egencache -j "${JOBS}" --cache-dir /var/cache/pgo-egencache --repo gentoo --repositories-configuration '[gentoo]
  location = /mnt/packages-tree/gentoo' --update-use-local-desc
}

fullupdate_database(){
  cd /mnt/packages-tree/gentoo/ || exit 1
  /go/src/soko/bin/soko --fullupdate
}


update_repository
update_md5cache
fullupdate_database
