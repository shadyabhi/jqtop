package main

import (
	"fmt"

	"github.com/shadyabhi/jqtop/argparser"
)

func parseFilters(s string) ([]filter, error) {
	if s == "" {
		return []filter{}, nil
	}
	filtersAST, err := argparser.ParseFilters(s)
	if err != nil {
		return []filter{}, fmt.Errorf("Error parsing filter, error: %s", err)
	}

	filters := []filter{}
	validFunctions := getFieldValues(filterFunctions)
	for _, f := range filtersAST.Filters {
		if !sliceContains(validFunctions, *f.Name) {
			return []filter{}, fmt.Errorf("Error parsing filter, invalid action: %s", *f.Name)
		}
		var negate bool

		// For now, we only worry about "!"
		// hence named negate
		if f.Unaryop == nil {
			negate = false
		} else {
			switch op := *f.Unaryop; op {
			case "!":
				negate = true
			}
		}

		args := []string{}
		for i := range f.Args {
			if i == 0 {
				args = append(args, *f.Args[i].Variable)
			} else {
				args = append(args, *f.Args[i].String)
			}
		}
		eachFilter := filter{
			Negate:   negate,
			Function: *f.Name,
			Args:     args,
		}
		filters = append(filters, eachFilter)
	}
	return filters, nil
}
