// nolint: golint
package argparser

import (
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var customLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(?P<Semicolon>;)` +
		`|(?P<Comma>,)` +
		`|(?P<At>@)` +
		`|(?P<Unaryop>[!])` +
		`|(?P<Equals>[=])` +
		`|(?P<Ident>[a-zA-Z][a-zA-Z_\d]*)` +
		`|(?P<String>"(?:\\.|[^"])*")` +
		`|(?P<Float>\d+(?:\.\d+)?)` +
		`|(?P<Lparen>[\(])` +
		`|(?P<Rparen>[\)])`,
))

// FieldExprs is used to parse fields
// Eg: field2(1, 2); field1; f3 = foo(1,2);
type FieldExprs struct {
	Exprs []*fieldExpr `parser:"{ @@ }"`
}

type fieldExpr struct {
	Assignment *assignment `parser:" @@ "`
	Expr       *expr       `parser:"| @@"`
}

type assignment struct {
	Variable string `parser:"@Ident \"=\""`
	Expr     *expr  `parser:"@@"`
}

type expr struct {
	Function *function `parser:"@@"`
	Term     *term     `parser:"| @@ { \";\" }"`
}

// FilterExprs is used to parse filters
// Eg: !contains(cquuc, "foo") regex(domain, "perf.linkedin.com")
type FilterExprs struct {
	Filters []*function `parser:"{ @@ }"`
}

type function struct {
	Unaryop *string `parser:"{ @Unaryop }"`
	Name    *string `parser:"@Ident"`
	Args    []*term `parser:"  \"(\" [ @@ { \",\" @@ } ] \")\" { \";\" }"`
}

type term struct {
	Variable *string  `parser:"@Ident"`
	String   *string  `parser:"| @String"`
	Number   *float64 `parser:"| @Float"`
}

// ParseFields parses "s" based on a the grammer described
// by "FieldExprs"
func ParseFields(s string) (*FieldExprs, error) {
	parser, err := participle.Build(
		&FieldExprs{},
		participle.Lexer(customLexer),
		participle.Unquote(customLexer, "String"),
		participle.UseLookahead(),
	)
	if err != nil {
		return &FieldExprs{}, err
	}
	myAST := &FieldExprs{}
	err = parser.Parse(strings.NewReader(s), myAST)
	if err != nil {
		return &FieldExprs{}, err
	}
	return myAST, nil
}

// ParseFilters parses "s" based on a the grammer described
// by "FilterExprs"
func ParseFilters(s string) (*FilterExprs, error) {
	parser, err := participle.Build(&FilterExprs{}, participle.Lexer(customLexer), participle.Unquote(customLexer, "String"))
	if err != nil {
		return &FilterExprs{}, err
	}
	myAST := &FilterExprs{}
	err = parser.Parse(strings.NewReader(s), myAST)
	if err != nil {
		return &FilterExprs{}, err
	}
	return myAST, nil
}
