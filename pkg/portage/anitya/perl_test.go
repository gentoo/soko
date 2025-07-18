package anitya

import "testing"

func TestPerlWrapper(t *testing.T) {
	pv, err := NewPerlVersion()
	if err == errNoPerlVersionScript {
		t.Skipf("Skipping test: %v", err)
	} else if err != nil {
		t.Fatalf("Failed to create PerlVersion: %v", err)
	}
	defer func() {
		err := pv.cmd.Process.Kill()
		if err != nil {
			t.Errorf("Failed to kill PerlVersion process: %v", err)
		}
	}()

	t.Run("Bad cases", func(t *testing.T) {
		testCases := []string{
			"invalid-version",
			"1.0a6",
			"1.02ii",
		}
		for _, input := range testCases {
			t.Run(input, func(t *testing.T) {
				_, err := pv.Query(input)
				if err == nil {
					t.Errorf("Expected error for input %q, got nil", input)
				}
			})
		}
	})

	// we also expect the handler to not crash after bad inputs

	t.Run("Good cases", func(t *testing.T) {
		testCases := map[string]string{
			"":         "",
			"0":        "0.0.0",
			"02":       "2.0.0",
			"1.0.0":    "1.0.0",
			"1.000555": "1.0.555",
		}

		for input, expected := range testCases {
			t.Run(input, func(t *testing.T) {
				result, err := pv.Query(input)
				if err != nil {
					t.Errorf("Query failed: %v", err)
					return
				}
				if result != expected {
					t.Errorf("Expected %q, got %q", expected, result)
				}
			})
		}
	})
}
