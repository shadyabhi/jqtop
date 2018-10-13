package jqtop

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hpcloud/tail"
	randomdata "github.com/shadyabhi/go-randomdata"
)

func linesGenerator(linesChan chan *tail.Line, n int) {
	for i := 0; i < n; i++ {
		line := tail.NewLine(fmt.Sprintf("{\"ttms\": %d, \"code\": %d, \"domain\": \"%s/%s\"}\n",
			randomdata.Number(100), randomdata.Number(599), randomdata.Domain(), randomdata.Noun()))
		linesChan <- line
	}
	close(linesChan)
}

func BenchmarkUpdateMapWithMutex(b *testing.B) {
	var sharedMap struct {
		M    map[string]string
		Lock sync.RWMutex
	}
	sharedMap.M = make(map[string]string)

	b.Run("Update a map with mutex", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			sharedMap.Lock.Lock()
			sharedMap.M["foo"] = "bar"
			sharedMap.Lock.Unlock()
		}
	})

	nGoRoutines := []int{1, 10, 100, 500, 1000, 5000, 10000}

	for _, nR := range nGoRoutines {

		b.Run(fmt.Sprintf("Update a map with mutex (%d GoRoutines)", nR), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var wg sync.WaitGroup

				wg.Add(nR)
				for i := 0; i < nR; i++ {
					go func() {
						defer wg.Done()
						sharedMap.Lock.Lock()
						sharedMap.M["foo"] = "bar"
						sharedMap.Lock.Unlock()
					}()
				}
				wg.Wait()
			}
		})
	}
}

func BenchmarkJqtop(b *testing.B) {

	nLines := []int{1, 10000, 100000, 1000000}

	// Setup DumpCounters goroutine
	buf := make([]byte, 1000)
	outStream := bytes.NewBuffer(buf)
	go DumpCounters(outStream)
	// go DumpCounters(nil)

	// Common args
	args.Interval = 1000

	args.Fields = "field_doesnt_exist"
	runJqtopWithArgs(b, "Get stats for non-existent field", args, nLines)

	args.Fields = "domain"
	runJqtopWithArgs(b, "Get stats for one field", args, nLines)

	args.Fields = "domain"
	args.Filters = "contains(domain, \".com\")"
	runJqtopWithArgs(b, "Get stats for one field with basic filter (domain contains .com)", args, nLines)

	args.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	args.Filters = ""
	runJqtopWithArgs(b, "Get stats for creating new field via regex", args, nLines)

	args.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	args.Filters = "equals(domain_only, \"google.com\")"
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com", args, nLines)

	args.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	args.Filters = "equals(domain_only, \"google.com\")"
	args.ParallelProc = 4
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com(parallal = 4)", args, nLines)

	args.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	args.Filters = "equals(domain_only, \"google.com\")"
	args.ParallelProc = 12
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com(parallal = 12)", args, nLines)
}

func runJqtopWithArgs(b *testing.B, summary string, args Arguments, nLines []int) {
	var avgRuntimes []time.Duration

	for _, n := range nLines {
		var runTimes []int64
		b.Run(fmt.Sprintf("Process %d lines", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				linesChan := make(chan *tail.Line, n)
				// Generate
				linesGenerator(linesChan, n)

				// Consume
				now := time.Now()
				ProcessLines(linesChan)
				runTimes = append(runTimes, int64(time.Now().Sub(now)))
			}
		})

		var total int64
		for _, value := range runTimes {
			total = total + value
		}
		avgRunTime := time.Duration(total / int64(len(runTimes)))
		avgRuntimes = append(avgRuntimes, avgRunTime)
	}

	fmt.Printf("\nResults: %s\n", summary)
	fmt.Println("---------------------------------------------")
	for i, n := range nLines {
		fmt.Printf("Average time elapsed in processing %-10d lines: %-10s\n",
			n, avgRuntimes[i])
	}
	fmt.Printf("QPS :: %-10f\n\n", 1000000000*float64(nLines[len(nLines)-1])/float64(avgRuntimes[len(nLines)-1]))
}
