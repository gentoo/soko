package models

import (
	"fmt"
	"testing"
)

func TestVersionCompare(t *testing.T) {
	var tests = []struct {
		left, right string
		want        bool
	}{
		{"6.0", "5.0", true},
		{"5.0", "5", true},
		{"1.0-r1", "1.0-r0", true},
		{"1.0-r1", "1.0", true},
		{"999999999999999999999999999999", "999999999999999999999999999998", true},
		{"1.0.0", "1.0", true},
		{"1.0.0", "1.0b", true},
		{"1b", "1", true},
		{"1b_p1", "1_p1", true},
		{"1.1b", "1.1", true},
		{"12.2.5", "12.2b", true},
		{"4.0", "4.0", true},
		{"1.0", "1.0", true},
		{"1.0-r0", "1.0", true},
		{"1.0", "1.0-r0", true},
		{"1.0-r0", "1.0-r0", true},
		{"1.0-r1", "1.0-r1", true},
		{"4.0", "5.0", false},
		{"1.0_pre2", "1.0_p2", false},
		{"1.0_alpha2", "1.0_p2", false},
		{"1.0_alpha1", "1.0_beta1", false},
		{"1.0_beta3", "1.0_rc3", false},
		{"1.001000000000000000001", "1.001000000000000000002", false},
		{"1.00100000000", "1.0010000000000000001", false},
		{"999999999999999999999999999998", "999999999999999999999999999999", false},
		{"1.01", "1.1", false},
		{"1.0-r0", "1.0-r1", false},
		{"1.0", "1.0-r1", false},
		{"1.0", "1.0.0", false},
		{"1.0b", "1.0.0", false},
		{"1_p1", "1b_p1", false},
		{"1", "1b", false},
		{"1.1", "1.1b", false},
		{"12.2b", "12.2.5", false},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s.compare(%s)", tt.left, tt.right)
		t.Run(testname, func(t *testing.T) {
			left := Version{Version: tt.left}
			right := Version{Version: tt.right}
			// CompareTo is really >=, not Equality Comparison.
			ret := left.CompareTo(right)
			if ret != tt.want {
				t.Errorf("got %t, want %t", ret, tt.want)
			}
		})
	}
}
