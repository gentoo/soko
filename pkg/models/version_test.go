package models

import (
	"fmt"
	"testing"
)

func TestVersion_GreaterThan(t *testing.T) {
	var tests = []struct {
		left, right string
		want        bool
	}{
		// left version should be greater than the right version
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
		{"0.7.8_p0-r100", "0.7.8_p0-r2", true},

		// left version should be equal to the right version
		{"4.0", "4.0", false},
		{"1.0", "1.0", false},
		{"1.0-r0", "1.0", false},
		{"1.0", "1.0-r0", false},
		{"1.0-r0", "1.0-r0", false},
		{"1.0-r1", "1.0-r1", false},

		// left version should be less than the right version
		{"4.0", "5.0", false},
		{"5", "5.0", false},
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
		testname := fmt.Sprintf("%s.greaterThan(%s)", tt.left, tt.right)
		t.Run(testname, func(t *testing.T) {
			left := Version{Version: tt.left}
			right := Version{Version: tt.right}
			ret := left.GreaterThan(right)
			if ret != tt.want {
				t.Errorf("got %t, want %t", ret, tt.want)
			}
		})
	}
}

func TestVersion_SmallerThan(t *testing.T) {
	var tests = []struct {
		left, right string
		want        bool
	}{
		// left version should be greater than the right version
		{"6.0", "5.0", false},
		{"5.0", "5", false},
		{"1.0-r1", "1.0-r0", false},
		{"1.0-r1", "1.0", false},
		{"999999999999999999999999999999", "999999999999999999999999999998", false},
		{"1.0.0", "1.0", false},
		{"1.0.0", "1.0b", false},
		{"1b", "1", false},
		{"1b_p1", "1_p1", false},
		{"1.1b", "1.1", false},
		{"12.2.5", "12.2b", false},

		// left version should be equal to the right version
		{"4.0", "4.0", false},
		{"1.0", "1.0", false},
		{"1.0-r0", "1.0", false},
		{"1.0", "1.0-r0", false},
		{"1.0-r0", "1.0-r0", false},
		{"1.0-r1", "1.0-r1", false},

		// left version should be less than the right version
		{"4.0", "5.0", true},
		{"5", "5.0", true},
		{"1.0_pre2", "1.0_p2", true},
		{"1.0_alpha2", "1.0_p2", true},
		{"1.0_alpha1", "1.0_beta1", true},
		{"1.0_beta3", "1.0_rc3", true},
		{"1.001000000000000000001", "1.001000000000000000002", true},
		{"1.00100000000", "1.0010000000000000001", true},
		{"999999999999999999999999999998", "999999999999999999999999999999", true},
		{"1.01", "1.1", true},
		{"1.0-r0", "1.0-r1", true},
		{"1.0", "1.0-r1", true},
		{"1.0", "1.0.0", true},
		{"1.0b", "1.0.0", true},
		{"1_p1", "1b_p1", true},
		{"1", "1b", true},
		{"1.1", "1.1b", true},
		{"12.2b", "12.2.5", true},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s.lessThan(%s)", tt.left, tt.right)
		t.Run(testname, func(t *testing.T) {
			left := Version{Version: tt.left}
			right := Version{Version: tt.right}
			ret := left.SmallerThan(right)
			if ret != tt.want {
				t.Errorf("got %t, want %t", ret, tt.want)
			}
		})
	}
}

func TestVersion_EqualTo(t *testing.T) {
	var tests = []struct {
		left, right string
		want        bool
	}{
		// left version should be greater than the right version
		{"6.0", "5.0", false},
		{"5.0", "5", false},
		{"1.0-r1", "1.0-r0", false},
		{"1.0-r1", "1.0", false},
		{"999999999999999999999999999999", "999999999999999999999999999998", false},
		{"1.0.0", "1.0", false},
		{"1.0.0", "1.0b", false},
		{"1b", "1", false},
		{"1b_p1", "1_p1", false},
		{"1.1b", "1.1", false},
		{"12.2.5", "12.2b", false},

		// left version should be equal to the right version
		{"4.0", "4.0", true},
		{"1.0", "1.0", true},
		{"1.0-r0", "1.0", true},
		{"1.0", "1.0-r0", true},
		{"1.0-r0", "1.0-r0", true},
		{"1.0-r1", "1.0-r1", true},

		// left version should be less than the right version
		{"4.0", "5.0", false},
		{"5", "5.0", false},
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
		testname := fmt.Sprintf("%s.equalTo(%s)", tt.left, tt.right)
		t.Run(testname, func(t *testing.T) {
			left := Version{Version: tt.left}
			right := Version{Version: tt.right}
			ret := left.EqualTo(right)
			if ret != tt.want {
				t.Errorf("got %t, want %t", ret, tt.want)
			}
		})
	}
}
