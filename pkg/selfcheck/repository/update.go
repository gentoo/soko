// SPDX-License-Identifier: GPL-2.0-only
// Update the portage data in the database

package repository

import (
	"soko/pkg/selfcheck/portage"
)

func Import() {
	for _, path := range AllFiles() {
		repository.UpdateVersion(path)
		repository.UpdatePackage(path)
		repository.UpdateCategory(path)

		repository.UpdateUse(path)
		repository.UpdateMask(path)
		repository.UpdateArch(path)
	}
}
