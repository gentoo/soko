// Contains functions to import the list of arches into the database

package repository

// isArchList checks whether the path
// points to a arch.list file
func isArchList(path string) bool {
	return path == "profiles/arch.list"
}

func UpdateArch(line string) {
	//TODO
}
