package main

import "log"

func main() {
	parseArgs()

	t, err := tailThis(args.File)
	if err != nil {
		log.Fatalf("Unable to tail: %s", err)
	}

	go dumpCounters()

	processLines(t.Lines)
}
