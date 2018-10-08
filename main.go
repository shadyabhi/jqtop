package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := parseArgs(); err != nil {
		log.Fatalf("Error parsing cmdline args: %s", err)
	}

	t, err := tailThis(args.File)
	if err != nil {
		log.Fatalf("Unable to tail: %s", err)
	}

	go dumpCounters()

	processLines(t.Lines)
}
