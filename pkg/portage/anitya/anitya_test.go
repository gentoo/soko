// SPDX-License-Identifier: GPL-2.0-only
package anitya

import "testing"

func TestLatestVersion(t *testing.T) {
	tests := []struct {
		version, stable, want string
	}{
		{"1.0.0b2", "", "1.0.0_beta2"},
		{"3.1.0a1", "3.1.0a1", "3.1.0_alpha1"},
		{"2.0rc3", "", "2.0_rc3"},
		{"2.0-beta.1", "", "2.0_beta1"},
		{"1.1.1a", "", "1.1.1a"},
		{"1.0b", "", "1.0b"},
	}
	for _, tt := range tests {
		p := ApiPackage{Version: tt.version, StableVersion: tt.stable}
		if got := p.LatestVersion(); got != tt.want {
			t.Errorf("LatestVersion(%q, %q) = %q, want %q", tt.version, tt.stable, got, tt.want)
		}
	}
}
