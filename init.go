package jqtop

import (
	"time"

	"github.com/paulbellamy/ratecounter"
)

func init() {
	// Defaults
	Config.Interval = 1
	Config.MaxResult = 10
	Config.ParallelProc = 4

	parseErrors = ratecounter.NewRateCounter(time.Duration(Config.Interval) * time.Second)
}
