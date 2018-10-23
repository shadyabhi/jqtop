package jqtop

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/struCoder/pidusage"
)

//counters is used for sorting
type sortedCounters struct {
	Name       string
	Value      int64
	Percentage float64
}

type printerStats struct {
	sysinfo     *pidusage.SysInfo
	timeElapsed time.Duration
}

func getSortedCounters(fMap *map[string]int, fName string, lastCounters *[]map[string]int64) (ss []sortedCounters) {
	countersSlice.RLock()
	r := countersSlice.counters[(*fMap)[fName]]
	countersSlice.RUnlock()

	var counters []sortedCounters

	// Sort
	var total float64
	r.Each(func(name string, i interface{}) {
		c := i.(*metrics.StandardCounter)
		rate := c.Count() - (*lastCounters)[(*fMap)[fName]][name]
		(*lastCounters)[(*fMap)[fName]][name] = c.Count()
		counters = append(counters, sortedCounters{name, rate, 0})
		total += float64(rate)
	})
	sort.Slice(counters, func(i, j int) bool {
		return counters[i].Value > counters[j].Value
	})

	// Get percentages
	for i := range counters {
		counters[i].Percentage = float64(counters[i].Value) / total * 100
	}

	// Return subset if needed
	if Config.MaxResult < 0 {
		return counters
	}
	if len(counters) > Config.MaxResult {
		return counters[:Config.MaxResult]
	}
	return counters
}

func printCounters(out io.Writer, counters map[string][]sortedCounters, stats printerStats) {
	if Config.Clearscreen {
		fmt.Println("\033[H\033[2J")
	}
	fmt.Fprintf(out, "✖ Parse error rate: %d, CPU Usage: %.2f%%, Mem: %.2fMB, Processing Time: %s\n",
		parseErrors.Rate(), stats.sysinfo.CPU, stats.sysinfo.Memory/1024.0/1024.0, stats.timeElapsed)

	for _, fieldName := range getFieldsInOrder(Config.Fields) {
		fmt.Fprintf(out, "➤ %s\n", fieldName)
		indent := "└──"
		for _, counterData := range counters[fieldName] {
			fmt.Fprintf(out, "%s %4s [%.1f%%]: %s\n",
				indent, strconv.FormatInt(counterData.Value, 10), counterData.Percentage, counterData.Name)
		}
		fmt.Fprintln(out, "")
	}
}

// DumpCounters is a function to print stats to io.Writer
// nil io.Write is stdout
func DumpCounters(out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	// Fields
	allFields, err := extractFields(Config.Fields)
	if err != nil {
		logrus.Fatalf("Error parsing fields, existing")
	}

	fMap := getFieldIndexMap(allFields.FieldsInOrder)
	// Holds last value (avoid timers)
	// Array of map[fieldname](counterValue)
	var lastCounters []map[string]int64
	initLastCounters(&lastCounters, allFields)

	ticker := time.NewTicker(time.Millisecond * time.Duration(Config.Interval))
	for range ticker.C {
		now := time.Now()
		stats := printerStats{}
		sysinfo, _ := getCPUUsage()
		stats.sysinfo = sysinfo

		counters := make(map[string][]sortedCounters)
		for _, fieldName := range getFieldsInOrder(Config.Fields) {
			// ss := getSortedCounters(countersMap.counters[fieldName])
			ss := getSortedCounters(&fMap, fieldName, &lastCounters)
			counters[fieldName] = ss
		}
		timeElapsed := time.Since(now).Round(time.Millisecond)
		stats.timeElapsed = timeElapsed
		printCounters(out, counters, stats)
	}
}

func initLastCounters(counters *[]map[string]int64, allFields Fields) {
	*counters = make([]map[string]int64, len(allFields.FieldsInOrder))
	for i := range allFields.FieldsInOrder {
		(*counters)[i] = make(map[string]int64, 0)
	}
}

func getCPUUsage() (sysInfo *pidusage.SysInfo, err error) {
	sysInfo, err = pidusage.GetStat(os.Getpid())
	if err != nil {
		return &pidusage.SysInfo{}, errors.Wrap(err, "Error getting CPU usage")
	}
	return sysInfo, nil
}
