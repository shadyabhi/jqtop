package main

import (
	"fmt"
	"go/parser"
	"jqtop/argparser"

	"go/ast"

	arg "github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
)

// Accepts command line arguments of the form "head.tail"
type complexField struct {
	NewField string
	Fname    string
	Args     []string
}

var complexFields map[string]*complexField
var simpleFields []string

type filter struct {
	Negate   bool
	Function string
	Pos      int
	// We can have more than 2 in future
	Args []string
}

func (f *complexField) UnmarshalText(b []byte) error {
	s := string(b)

	myAST, err := argparser.ParseFields(s)
	if err != nil {
		return fmt.Errorf("Invalid format for fields: %s", err)
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
			complexFields[expr.Assignment.Variable] = &complexField{expr.Assignment.Variable, *expr.Assignment.Expr.Function.Name, args}
		}
	}
	return nil
}

func getAST(s string) (ast.Expr, error) {
	ast, err := parser.ParseExpr(s)
	if err != nil {
		return nil, err
	}

	return ast, nil
}

var args struct {
	File      string `arg:"required"`
	Interval  int    `arg:"-i"`
	MaxResult int    `arg:"-m"`
	Verbose   bool   `arg:"-v"`

	// package doesn't support *[]complexField so
	// we create *complexFieldSlice and use
	// complexFields.Fields in our code
	ComplexField *complexField `arg:"-F,separate"`
	// Same as ComplexField
	Filter string `arg:"separate"`
}

// parseArgs parses args and validates
func parseArgs() {
	// Defaults are defined in init.go so they
	// can be used in tests too.
	arg.MustParse(&args)
	if args.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Debugf("Parsed following arguments: %+v", args)
}
