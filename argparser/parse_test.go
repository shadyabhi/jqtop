package argparser

import (
	"testing"
)

func TestParseFields(t *testing.T) {
	myAST, err := ParseFields(`field1; field2; field3 = func1("a");`)
	if err != nil {
		t.Errorf("Error parsing AST: %s", err)
	}
	if len(myAST.Exprs) != 3 {
		t.Errorf("All fields were not parsed")
	}
}

func TestParseFilters(t *testing.T) {
	myAST, err := ParseFilters(`!contains(a,21); regex("a", "b"); foobar("a");`)
	if err != nil {
		t.Errorf("Error parsing AST: %s", err)
	}
	if len(myAST.Filters) != 3 {
		t.Errorf("myAST doesn't have enough filters")
	}
	if *myAST.Filters[0].Name != "contains" {
		t.Errorf("myAST doesn't have first function as \"contains\"")
	}
	if *myAST.Filters[1].Name != "regex" {
		t.Errorf("myAST doesn't have second function as \"regex\"")
	}
	if *myAST.Filters[2].Args[0].String != "a" {
		t.Errorf("myAST doesn't have third function's argument as \"a\"")
	}
}
