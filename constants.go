package main

// Valid string options for filterField
var fieldActions = struct {
	Exclude string
	Include string
}{"exclude", "include"}

var filterFunctions = struct {
	Contains string
	Regex    string
	Equal    string
}{"contains", "regex", "equals"}
