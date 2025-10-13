// SPDX-License-Identifier: GPL-2.0-only
package anitya

// dev-perl packages have a special versioning scheme in Gentoo
// https://wiki.gentoo.org/wiki/Project:Perl/Version-Scheme

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var errNoPerlVersionScript = errors.New("perl-version.pl not found in standard locations")

type PerlVersion struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       *bufio.Reader
	versionRegex *regexp.Regexp
}

func NewPerlVersion() (*PerlVersion, error) {
	scriptPaths := []string{
		"/usr/share/pkgcheck/perl-version.pl",
		"/usr/local/share/pkgcheck/perl-version.pl",
	}
	var script string
	for _, p := range scriptPaths {
		if _, err := os.Stat(p); err == nil {
			script = p
			break
		}
	}
	if script == "" {
		return nil, errNoPerlVersionScript
	}

	cmd := exec.Command("perl", script)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start perl-version.pl: %w", err)
	}

	stdout := bufio.NewReader(stdoutPipe)
	resp, err := stdout.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read 'ready' from perl-version.pl: %w", err)
	} else if strings.TrimSpace(resp) != "ready" {
		return nil, fmt.Errorf("unexpected response from perl-version.pl: %q", resp)
	}

	return &PerlVersion{
		cmd:          cmd,
		stdin:        stdin,
		stdout:       stdout,
		versionRegex: regexp.MustCompilePOSIX(`^[0-9]+(\.[0-9]+)*$`),
	}, nil
}

func (pv *PerlVersion) Query(line string) (string, error) {
	if line == "" {
		return "", nil
	} else if !pv.versionRegex.MatchString(line) {
		return "", fmt.Errorf("invalid version format: %s", line)
	}
	_, err := pv.stdin.Write([]byte(line + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to write to perl-version.pl: %w", err)
	}
	resp, err := pv.stdout.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from perl-version.pl: %w", err)
	}
	return strings.TrimSpace(resp), nil
}

func (pv *PerlVersion) Close() error {
	pv.stdin.Close()
	return pv.cmd.Wait()
}
