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

// field2(1, 2); field1; f3 = foo(1,2);
type FieldExprs struct {
	Exprs []*FieldExpr `{ @@ }`
}

type FieldExpr struct {
	Assignment *Assignment ` @@ `
	Expr       *Expr       `| @@`
}

type Assignment struct {
	Variable string `@Ident "="`
	Expr     *Expr  `@@`
}

type Expr struct {
	Function *Function `  @@`
	Term     *Term     `| @@ { ";" }`
}

// !contains(cquuc, "foo") regex(domain, "perf.linkedin.com")
type FilterExprs struct {
	Filters []*Function `parser:"{ @@ }"`
}

type Function struct {
	Unaryop *string `parser:"{ @Unaryop }"`
	Name    *string `@Ident`
	Args    []*Term `"(" [ @@ { "," @@ } ] ")" { ";" }`
}

type Term struct {
	Variable *string  `@Ident`
	String   *string  `| @String`
	Number   *float64 `| @Float`
}

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
	// repr.Println(myAST, repr.Indent("  "), repr.OmitEmpty(true))
	return myAST, nil
}

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
	// repr.Println(myAST, repr.Indent("  "), repr.OmitEmpty(true))
	// spew.Dump(myAST)
	return myAST, nil
}
