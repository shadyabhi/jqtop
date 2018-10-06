package main

import (
	"fmt"

	"github.com/shadyabhi/jqtop/argparser"
)

type derivedField struct {
	NewField string
	Fname    string
	Args     []string
}

// Fields contains parsed fields that are sent
type Fields struct {
	SimpleFields  []string
	DerivedFields map[string]*derivedField
}

func extractFields(s string) (Fields, error) {
	allFields := Fields{}
	simpleFields := []string{}
	derivedFields := make(map[string]*derivedField)

	myAST, err := argparser.ParseFields(s)
	if err != nil {
		return allFields, fmt.Errorf("Invalid format for fields: %s", err)
	}

	for _, expr := range myAST.Exprs {
		if expr.Assignment == nil {
			// Simple field
			simpleFields = append(simpleFields, *expr.Expr.Term.Variable)
		} else {
			// Complex field
			// TODO: Convert to function
			args := []string{}
			for i := range expr.Assignment.Expr.Function.Args {
				if i == 0 {
					args = append(args, *expr.Assignment.Expr.Function.Args[i].Variable)
				} else {
					args = append(args, *expr.Assignment.Expr.Function.Args[i].String)
				}
			}
			derivedFields[expr.Assignment.Variable] = &derivedField{expr.Assignment.Variable, *expr.Assignment.Expr.Function.Name, args}
		}
	}

	allFields = Fields{
		SimpleFields:  simpleFields,
		DerivedFields: derivedFields,
	}
	return allFields, nil
}
