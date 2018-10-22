package jqtop

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

//counterData is used for sorting
type counterData struct {
	Key   string
	Value int64
}

//counters is used for sorting
type sortedCounters struct {
	Key   string
	Value int64
}

func getSortedCounters(fMap *map[string]int, fName string, lastCounters *[]map[string]int64) (ss []sortedCounters) {
	r := countersSlice.counters[(*fMap)[fName]]
	r.Each(func(name string, i interface{}) {
		c := i.(*metrics.StandardCounter)
		rate := c.Count() - (*lastCounters)[(*fMap)[fName]][name]
		(*lastCounters)[(*fMap)[fName]][name] = c.Count()
		ss = append(ss, sortedCounters{name, rate})
	})
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	if Config.MaxResult < 0 {
		return ss
	}
	if len(ss) > Config.MaxResult {
		return ss[:Config.MaxResult]
	}
	return ss
}

func printCounter(out io.Writer, fieldName string, ss []sortedCounters) {
	fmt.Fprintf(out, "➤ %s\n", fieldName)
	indent := "└──"
	for _, counterData := range ss {
		fmt.Fprintf(out, "%s %4s: %s\n", indent, strconv.FormatInt(counterData.Value, 10), counterData.Key)
	}
	fmt.Fprintln(out, "")
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
		if Config.Clearscreen {
			fmt.Println("\033[H\033[2J")
		}

		fmt.Fprintf(out, "✖ Parse error rate: %d\n", parseErrors.Rate())

		// countersMap.mu.RLock()
		for _, fieldName := range getFieldsInOrder(Config.Fields) {
			// ss := getSortedCounters(countersMap.counters[fieldName])
			ss := getSortedCounters(&fMap, fieldName, &lastCounters)
			printCounter(out, fieldName, ss)
		}
		// countersMap.mu.RUnlock()
	}
}

func initLastCounters(counters *[]map[string]int64, allFields Fields) {
	*counters = make([]map[string]int64, len(allFields.FieldsInOrder))
	for i := range allFields.FieldsInOrder {
		(*counters)[i] = make(map[string]int64, 0)
	}
}
