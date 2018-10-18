package jqtop

import (
	"time"

	"github.com/paulbellamy/ratecounter"
)

func init() {
	// Defaults
	Args.Interval = 1
	Args.MaxResult = 10
	Args.ParallelProc = 4

	parseErrors = ratecounter.NewRateCounter(time.Duration(Args.Interval) * time.Second)
}
