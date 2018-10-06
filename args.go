package main

import (
	arg "github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"
)

var args struct {
	File        string `arg:"required" help:"Path to file that will be read"`
	Interval    int    `arg:"-i" help:"Interval at which stats are calculated"`
	MaxResult   int    `arg:"-m" help:"Max results to show"`
	Verbose     bool   `arg:"-v"`
	Clearscreen bool   `arg:"-c" help:"Clear screen each time stats are shown"`

	Fields  string `arg:"separate" help:"Fields that need to shown for stats"`
	Filters string `arg:"separate" help:"Filters to filter lines that'll be processed"`
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
