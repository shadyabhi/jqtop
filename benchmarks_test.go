package jqtop

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hpcloud/tail"
	"github.com/paulbellamy/ratecounter"
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
	Config.Interval = 1000
	Config.ParallelProc = 1

	Config.Fields = "field_doesnt_exist"
	runJqtopWithArgs(b, "Get stats for non-existent field", Config, nLines)

	Config.Fields = "domain"
	runJqtopWithArgs(b, "Get stats for one field", Config, nLines)

	Config.Fields = "domain"
	Config.Filters = "contains(domain, \".com\")"
	runJqtopWithArgs(b, "Get stats for one field with basic filter (domain contains .com)", Config, nLines)

	Config.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	Config.Filters = ""
	runJqtopWithArgs(b, "Get stats for creating new field via regex", Config, nLines)

	Config.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	Config.Filters = "equals(domain_only, \"google.com\")"
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com", Config, nLines)

	Config.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	Config.Filters = "equals(domain_only, \"google.com\")"
	Config.ParallelProc = 4
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com(parallal = 4)", Config, nLines)

	Config.Fields = "domain_only = regex_capture(domain, \"(.*)/\")"
	Config.Filters = "equals(domain_only, \"google.com\")"
	Config.ParallelProc = 12
	runJqtopWithArgs(b, "Get stats for creating new field via regex and filter only google.com(parallal = 12)", Config, nLines)
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

func BenchmarkRateCounterIncr(b *testing.B) {
	// Focus is to check how Rate Counter performs
	// because of timers

	n := 10000
	counters := make([]*ratecounter.RateCounter, n)
	for i := 0; i < n; i++ {
		counters[i] = ratecounter.NewRateCounter(1 * time.Millisecond)
	}

	for i := 0; i < b.N; i++ {
		for i := 0; i < n; i++ {
			counters[i].Incr(1)
		}
	}
}

func BenchmarkAtomicIncr(b *testing.B) {
	var counter ratecounter.Counter

	b.Run("No goroutines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			counter.Incr(1)
		}
	})
	b.Run("With goroutines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			go func() {
				counter.Incr(1)
			}()
		}
	})
}

func BenchmarkMutexIncr(b *testing.B) {
	var counter struct {
		c int64
		sync.Mutex
	}

	b.Run("No goroutines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			counter.Lock()
			counter.c = counter.c + 1
			counter.Unlock()
		}

	})
	b.Run("With goroutines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			go func() {
				counter.Lock()
				counter.c = counter.c + 1
				counter.Unlock()
			}()
		}

	})
}
