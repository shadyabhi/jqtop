package jqtop

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
	SimpleFields   []string
	DerivedFields  map[string]*derivedField
	FieldsInOrder  []string
	FieldsIndexMap map[string]int
}

func getFieldsInOrder(s string) (allFields []string) {
	myAST, _ := argparser.ParseFields(s)
	for _, expr := range myAST.Exprs {
		if expr.Assignment != nil {
			allFields = append(allFields, expr.Assignment.Variable)
		} else {
			allFields = append(allFields, *expr.Expr.Term.Variable)
		}
	}
	return allFields
}

func extractFields(s string) (allFields Fields, err error) {
	simpleFields := []string{}
	fieldsInOrder := []string{}
	derivedFields := make(map[string]*derivedField)

	myAST, err := argparser.ParseFields(s)
	if err != nil {
		return allFields, fmt.Errorf("Invalid format for fields: %s", err)
	}

	for _, expr := range myAST.Exprs {
		if expr.Assignment == nil {
			// Simple field
			f := *expr.Expr.Term.Variable
			simpleFields = append(simpleFields, f)
			fieldsInOrder = append(fieldsInOrder, f)
		} else {
			// Complex field
			fieldsInOrder = append(fieldsInOrder, expr.Assignment.Variable)
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

	fMap := getFieldIndexMap(fieldsInOrder)
	allFields = Fields{
		SimpleFields:   simpleFields,
		DerivedFields:  derivedFields,
		FieldsInOrder:  fieldsInOrder,
		FieldsIndexMap: fMap,
	}

	return allFields, nil
}

func getFieldIndexMap(fieldsInOrder []string) (m map[string]int) {
	m = make(map[string]int)
	for i, v := range fieldsInOrder {
		m[v] = i
	}
	return m
}
