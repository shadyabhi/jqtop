package jqtop

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/paulbellamy/ratecounter"
	"github.com/sirupsen/logrus"
)

//counterData is used for sorting
type counterData struct {
	Key   string
	Value int64
}

// getSortedCounters takes counters map and returns a slice in decreasing order
// of ratecounter.RateCounter value
func getSortedCounters(counters map[string]*ratecounter.RateCounter) []counterData {
	var ss []counterData
	for k, counter := range counters {
		ss = append(ss, counterData{k, counter.Rate()})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	if Args.MaxResult < 0 {
		return ss
	}
	if len(ss) > Args.MaxResult {
		//Only show "top" rate
		return ss[:Args.MaxResult]
	}
	return ss
}

func printCounter(out io.Writer, fieldName string, ss []counterData) {
	fmt.Fprintf(out, "➤ %s\n", fieldName)
	indent := "└──"
	for _, counterData := range ss {
		fmt.Fprintf(out, "%s %4s: %s\n", indent, strconv.FormatInt(counterData.Value, 10), counterData.Key)
	}
	fmt.Fprintln(out, "")
}

// DumpCounters is a function to print stats to stdout
func DumpCounters(out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	ticker := time.NewTicker(time.Millisecond * time.Duration(Args.Interval))
	for range ticker.C {
		if Args.Clearscreen {
			fmt.Println("\033[H\033[2J")
		}

		fmt.Fprintf(out, "✖ Parse error rate: %d\n", parseErrors.Rate())
		count := 0

		countersMap.mu.RLock()
		for _, fieldName := range getFieldsInOrder(Args.Fields) {
			ss := getSortedCounters(countersMap.counters[fieldName])
			printCounter(out, fieldName, ss)
			count++
		}
		countersMap.mu.RUnlock()

		logrus.Debugf("dumpCounters: Total parsed counters = %d", count)
	}
}
