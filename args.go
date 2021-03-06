package jqtop

import (
	"os"

	arg "github.com/alexflint/go-arg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type described struct{}

func (described) Description() string {
	desc := `jqtop - realtime json analyzer

jqtop is a tool for processing newline delimited JSON stream of documents.
You can see stats of existing fields, create new fields via various functions
and filter documents that need to be processed.

Manual: https://github.com/shadyabhi/jqtop/wiki#manual

Sample usage:-

jqtop \
--file ./logfile \
--fields 'paths = regex_capture(request, "[A-Z]+? (.*?) "); http_method = regex_capture(request, "(.*?) ");' \
--filters 'equals(http_method, "POST")'

Filter functions:

* equals(field_name, "spammy.domain.com"
* contains(field_name, "needle")
* regex(field_name, "ignore.*")

Field functions:

* regex_capture(field_name, "//(.*?)/")

	`
	return desc
}

// Arguments describes the argument that the program receives
type Arguments struct {
	File         string  `arg:"" help:"Path to file that will be read"`
	Interval     float64 `arg:"-i" help:"Interval at which stats are calculated"`
	MaxResult    int     `arg:"-m" help:"Max results to show"`
	Verbose      bool    `arg:"-v"`
	Clearscreen  bool    `arg:"-c" help:"Clear screen each time stats are shown"`
	ParallelProc int     `arg:"-p" help:"Number of parallel processes for processing"`
	Iter         int     `arg:"-s" help:"Automatically stop after n number of iterations "`

	Fields  string `arg:"required,separate" help:"Fields that need to shown for stats"`
	Filters string `arg:"separate" help:"Filters to filter lines that'll be processed"`
}

// Config stores all arguments passed via CLI
var Config Arguments

// ParseArgs parses args and validates
func ParseArgs() error {
	// Defaults are defined in init.go so they
	// can be used in tests too.
	p, err := arg.NewParser(arg.Config{}, &Config, &described{})
	if err != nil {
		return errors.Wrap(err, "Error setting up parser:")
	}
	err = p.Parse(os.Args[1:])
	if err == arg.ErrHelp {
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	if err != nil {
		return errors.Wrap(err, "Error parsing cmdline arguments")
	}
	if Config.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Convert to seconds
	Config.Interval = 1000 * Config.Interval

	logrus.Debugf("Parsed following arguments: %+v", Config)
	return nil
}
