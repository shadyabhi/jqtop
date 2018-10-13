package jqtop

import (
	"testing"
)

func TestParseFilters(t *testing.T) {
	filters, err := parseFilters(`contains(cquuc, "coolsite.com"); !regex(cquuc, "www.*");`)
	if err != nil {
		t.Errorf("Error executing parseFilters: %s", err)
	}
	if len(filters) != 2 {
		t.Errorf("All filters were not parsed")
	}

	// Invalid filters
	_, err = parseFilters(`contains(`)
	if err == nil {
		t.Error("Expected error, got no error")
	}
}
