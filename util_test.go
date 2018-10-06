package main

import "testing"

func TestGetFieldValues(t *testing.T) {
	s := struct {
		Exclude string
		Include string
	}{"exclude", "include"}

	if !sliceEqual([]string{"exclude", "include"}, getFieldValues(s)) {
		t.Errorf("getFieldValues didn't fetch the right values")
	}

}

// sliceEqual tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestSliceContains(t *testing.T) {
	mySlice := []string{"hello", "world"}

	var testCases = []struct {
		in       string
		expected bool
	}{
		{"hello", true},
		{"not_exist", false},
	}
	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			if ret := sliceContains(mySlice, tt.in); ret != tt.expected {
				t.Errorf("TestSliceContains: For input: %s, we expected: %t", tt.in, tt.expected)
			}
		})
	}
}
