package jqtop

import (
	"time"

	"github.com/paulbellamy/ratecounter"
)

func init() {
	// Defaults
	args.Interval = 1
	args.MaxResult = 10
	args.ParallelProc = 1

	parseErrors = ratecounter.NewRateCounter(time.Duration(args.Interval) * time.Second)
}
