package jqtop

import (
	"reflect"
	"testing"
)

func TestGetFieldsInOrder(t *testing.T) {
	s := "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");"
	expected := []string{"cquuc", "host_with_protocol"}

	fields := getFieldsInOrder(s)
	for i, f := range fields {
		if expected[i] != f {
			t.Errorf("For fields %s, expected: %s, got: %s", s, expected[i], f)
		}
	}
}

func TestExtractFields(t *testing.T) {
	fieldsStr := "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");"
	allFields, err := extractFields(fieldsStr)
	t.Logf("Extracted fields\nFROM: %s\nGOT: %#v", fieldsStr, allFields)
	if err != nil {
		t.Errorf("Error parsinga valid field string")
	}
	derivedFields := make(map[string]*derivedField)
	derivedFields["host_with_protocol"] = &derivedField{
		NewField: "host_with_protocol",
		Fname:    "regex_capture",
		Args:     []string{"cquuc", "(.*?://.*?)/"},
	}
	expectedValue := Fields{
		SimpleFields:  []string{"cquuc"},
		DerivedFields: derivedFields,
		FieldsInOrder: []string{"cquuc", "host_with_protocol"},
	}
	if !reflect.DeepEqual(allFields, expectedValue) {
		t.Errorf("allFields struct was equal to what was expected")
	}
	// Invalid line
	fieldsStr = "field1; c = "
	_, err = extractFields(fieldsStr)
	if err == nil {
		t.Errorf("Didn't throw error for an invalid field string")
	}
}
