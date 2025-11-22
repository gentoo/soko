// SPDX-License-Identifier: GPL-2.0-only
package bugs

import (
	"strings"
	"testing"
)

func TestExtractAtom(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"dev-lang/perl-5.32.1 : some bug description", "dev-lang/perl"},
		{"app-editors/vim-8.2.3456-r1 : another bug", "app-editors/vim"},
		{"net-misc/curl-7.76.1  : curl bug", "net-misc/curl"},
		{"sys-libs/glibc-2.33  : glibc issue [gcc-16]", "sys-libs/glibc"},
		{"sys-apps/openrc-0.62.10 - tries to pollute read-only filesystem", "sys-apps/openrc"},
		{"sci-libs/opencascade[examples,inspector] must port to Qt6", "sci-libs/opencascade"},
		{"sys-kernel/installkernel: improve external kernel module rebulding", "sys-kernel/installkernel"},
		{"sys-kernel/dracut-108-r4 stable request", "sys-kernel/dracut"},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			summary, _, _ := strings.Cut(strings.TrimSpace(tc.input), " ")
			affectedPackage := versionSpecifierToPackageAtom(summary)
			if affectedPackage != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, affectedPackage)
			}
		})
	}
}
