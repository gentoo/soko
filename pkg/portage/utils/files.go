// Contains utility functions to read the content of files

package utils

import (
	"bufio"
	"os"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// FileExists checks whether the file
// at the given path does exist
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		return false
	}
}
