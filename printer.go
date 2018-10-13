package jqtop

import (
	"fmt"
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

	if args.MaxResult < 0 {
		return ss
	}
	if len(ss) > args.MaxResult {
		//Only show "top" rate
		return ss[:args.MaxResult]
	}
	return ss
}

func printCounter(fieldName string, ss []counterData) {
	fmt.Printf("➤ %s\n", fieldName)
	indent := "└──"
	for _, counterData := range ss {
		fmt.Printf("%s %4s: %s\n", indent, strconv.FormatInt(counterData.Value, 10), counterData.Key)
	}
	fmt.Println("")
}

// DumpCounters is a function to print stats to stdout
func DumpCounters() {
	ticker := time.NewTicker(time.Second * time.Duration(args.Interval))
	for range ticker.C {
		if args.Clearscreen {
			fmt.Println("\033[H\033[2J")
		}

		fmt.Printf("✖ Parse error rate: %d\n", parseErrors.Rate())
		count := 0

		countersMap.mu.RLock()
		for _, fieldName := range getFieldsInOrder(args.Fields) {
			ss := getSortedCounters(countersMap.counters[fieldName])
			printCounter(fieldName, ss)
			count++
		}
		countersMap.mu.RUnlock()

		logrus.Debugf("dumpCounters: Total parsed counters = %d", count)
	}
}
