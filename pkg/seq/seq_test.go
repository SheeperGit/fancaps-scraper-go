package seq

import (
	"reflect"
	"testing"
)

func TestParseSequenceString(t *testing.T) {
	max := 100
	oneToHundred, err := generateSequence(1, max, 1)
	if err != nil {
		t.Fatalf("failed to generate sequence: %v", err)
	}

	tests := []struct {
		input     string // Sequence string to parse.
		expected  []int  // Expected output.
		expectErr bool   // True if an error is expected from the given input.
	}{
		{"1-5", []int{1, 2, 3, 4, 5}, false},
		{"-5", []int{1, 2, 3, 4, 5}, false},
		{"1-", oneToHundred, false},
		{"1-13:3", []int{1, 4, 7, 10, 13}, false},
		{"1-4:2", []int{1, 3}, false},
		{"1", []int{1}, false},
		{"1, 13, 12, 19", []int{1, 12, 13, 19}, false},
		{"1-3, 2-4, 5-8:2, 6-8:2", []int{1, 2, 3, 4, 5, 6, 7, 8}, false},

		{"1--5", nil, true},
		{"foo-bar", nil, true},
		{"1-5:0", nil, true},
		{"5-1:", []int{}, true},
		{"3, 1-3, 23-2:0, 9-11", nil, true},
		{"1-101", nil, true},
		{"0-5", nil, true},
		{"0-101", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSequenceString(tt.input, max, false)

			if tt.expectErr {
				if err == nil {
					t.Errorf("parseSequenceString(%q) expected error but got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseSequenceString(%q) returned unexpected error: %v", tt.input, err)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseSequenceString(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
