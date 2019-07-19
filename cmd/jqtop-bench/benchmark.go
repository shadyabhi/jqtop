package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hpcloud/tail"
	"github.com/olekukonko/tablewriter"
	randomdata "github.com/shadyabhi/go-randomdata"
	"github.com/shadyabhi/jqtop"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/message"
)

func linesGenerator(linesChan chan *tail.Line, n int) {
	for i := 0; i < n; i++ {
		line := tail.NewLine(fmt.Sprintf("{\"ttms\": %d, \"code\": %d, \"domain\": \"%s/%s\"}\n",
			randomdata.Number(100), randomdata.Number(599), randomdata.Domain(), randomdata.Noun()))
		linesChan <- line
	}
	close(linesChan)
}

// BenchmarkJqtop benchmarks JQtop as an integration test
func BenchmarkJqtop(b *testing.B) {

	nLines := []int{1, 10000, 100000, 1000000}

	// Setup DumpCounters goroutine
	buf := make([]byte, 1000)
	outStream := bytes.NewBuffer(buf)
	go jqtop.DumpCounters(outStream, 0)

	benchArgs := make([][]string, 0)
	benchArgs = [][]string{
		{"1", "field_doesnt_exist", "", "Parse non-existent field"},
		{"1", "domain", "", "Parse one simple field"},
		{"1", "domain", "contains(domain, \".com\")", "Parse one simple field with basic filter (domain contains .com)"},
		{"1", "domain_only = regex_capture(domain, \"(.*)/\")", "", "Parse one derived field via regex"},
		{"6", "domain_only = regex_capture(domain, \"(.*)/\")", "", "Parse one derived field via regex"},
		{"1", "domain_only = regex_capture(domain, \"(.*)/\")", "equals(domain_only, \"google.com\")", "Parse one derived field with basic filter (domain equals google.com)"},
		{"4", "domain_only = regex_capture(domain, \"(.*)/\")", "equals(domain_only, \"google.com\")", "Parse one derived field with basic filter (domain equals google.com)"},
		{"12", "domain_only = regex_capture(domain, \"(.*)/\")", "equals(domain_only, \"google.com\")", "Parse one derived field with basic filter (domain equals google.com)"},
		{"12", "code; ttms; domain_only = regex_capture(domain, \"(.*)/\")", "equals(domain_only, \"google.com\")", "Parse multiple fields(simple/derived) with basic filter (domain equals google.com)"},
	}

	// Common jqtop.Args
	jqtop.Config.Interval = 1000

	tableData := make([][]string, len(benchArgs))
	for i := range tableData {
		tableData[i] = make([]string, 3)
	}

	for i, args := range benchArgs {
		parallelProcs, err := strconv.Atoi(args[0])
		if err != nil {
			logrus.Fatalf("Invalid input for bench test")
		}

		jqtop.Config.ParallelProc = parallelProcs
		jqtop.Config.Fields = args[1]
		jqtop.Config.Filters = args[2]
		summary := args[3]
		avgRunTime, qps := runJqtopWithArgs(b, summary, nLines)

		// Only show results table for highest number of lines
		// Last in nLines
		// Description
		tableData[i][0] = fmt.Sprintf("%d Go Routines: (%d lines): %s", parallelProcs, nLines[len(nLines)-1], summary)
		// Average Runtimes
		tableData[i][1] = fmt.Sprintf("%.2f seconds", float64(avgRunTime[len(nLines)-1])/float64(time.Second))
		// QPS
		p := message.NewPrinter(message.MatchLanguage("en"))
		tableData[i][2] = p.Sprintf("%.0f", qps[len(nLines)-1])

	}

	// Print table
	fmt.Printf("\n\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Description", "Time Elapsed", "QPS"})
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	for _, v := range tableData {
		table.Append(v)
	}
	table.Render()

}

func runJqtopWithArgs(b *testing.B, summary string, nLines []int) (avgRuntimes []time.Duration, qps []float64) {
	for _, n := range nLines {
		var runTimes []int64
		b.Run(fmt.Sprintf("Process %d lines", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				linesChan := make(chan *tail.Line, n)
				// Generate
				linesGenerator(linesChan, n)

				// Consume
				now := time.Now()
				jqtop.ProcessLines(linesChan)
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

	// Show some stats
	fmt.Printf("\nResults: %s\n", summary)
	fmt.Println("---------------------------------------------")
	for i, n := range nLines {
		fmt.Printf("Average time elapsed in processing %-10d lines: %-10s\n",
			n, avgRuntimes[i])
		qps = append(qps, 1000000000*float64(nLines[i])/float64(avgRuntimes[i]))
	}
	return avgRuntimes, qps
}
