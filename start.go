package jqtop

import (
	log "github.com/sirupsen/logrus"
)

func Start() {
	if err := ParseArgs(); err != nil {
		log.Fatalf("Error parsing cmdline args: %s", err)
	}

	t, err := TailFile(args.File)
	if err != nil {
		log.Fatalf("Unable to tail: %s", err)
	}

	go DumpCounters()

	ProcessLines(t.Lines)
}
