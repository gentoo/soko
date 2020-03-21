// Contains utility functions to execute commands and parse the output

package utils

import (
	"os/exec"
	"strings"
)

// Exec executes a given command with the given arguments in a given
// directory and returns a list of lines of the output of the command.
func Exec(dir string, command string, arg ...string) ([]string, error) {
	cmd := exec.Command(command, arg...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	return strings.Split(strings.TrimSpace(string(out)), "\n"), err
}
