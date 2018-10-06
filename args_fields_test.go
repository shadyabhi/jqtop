package main

import (
	"reflect"
	"testing"
)

func TestExtractFields(t *testing.T) {
	fieldsStr := "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");"
	allFields, err := extractFields(fieldsStr)
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
	}
	if reflect.DeepEqual(allFields, expectedValue) != true {
		t.Errorf("allFields struct was equal to what was expected")
	}
	// Invalid line
	fieldsStr = "field1; c = "
	allFields, err = extractFields(fieldsStr)
	if err == nil {
		t.Errorf("Didn't throw error for an invalid field string")
	}

}
