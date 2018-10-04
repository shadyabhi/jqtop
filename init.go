package main

import (
	"time"

	"github.com/paulbellamy/ratecounter"
)

func init() {
	complexFields = make(map[string]*complexField)

	// Defaults
	args.Interval = 1
	args.MaxResult = 10

	parseErrors = ratecounter.NewRateCounter(time.Duration(args.Interval) * time.Second)
}
