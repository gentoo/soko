// Contains functions to import package mask entries into the database

package repository

// isMask checks whether the path
// points to a package.mask file
func isMask(path string) bool {
	return path == "profiles/package.mask"
}

func UpdateMask(line string) {
	//TODO
}
