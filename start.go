package jqtop

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"

	"net/http"
	_ "net/http/pprof" // For profiling

	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
)

// Start function is called from outside for default behavior
func Start(outStream io.Writer) {
	if err := ParseArgs(); err != nil {
		log.Fatalf("Error parsing cmdline args: %s", err)
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	go DumpCounters(outStream)

	linesChan := make(chan *tail.Line)

	if err := setLinesChan(Config.File, linesChan); err != nil {
		log.Fatalf("Error reading lines, exiting with error: %s", err)
	}

	ProcessLines(linesChan)
}

func setLinesChan(filepath string, linesChan chan *tail.Line) (err error) {
	// File
	if filepath != "" {
		if err := tailF(linesChan, filepath); err != nil {
			return errors.Wrap(err, "Unable to tail file with error: ")
		}
		return nil
	}

	// Stdin
	log.Infof("Reading from stdin...\n")
	tailStdin(linesChan, os.Stdin)

	return nil
}

func tailStdin(linesChan chan *tail.Line, stdin *os.File) {
	go func() {
		stdinScanner := bufio.NewScanner(stdin)
		for {
			stdinScanner.Scan()
			line := tail.NewLine(stdinScanner.Text())
			linesChan <- line
		}
	}()
}

func tailF(linesChan chan *tail.Line, filepath string) error {
	t, err := TailFile(filepath)
	if err != nil {
		return errors.Wrap(err, "Unable to tail file with error: ")
	}
	go func() {
		for line := range t.Lines {
			linesChan <- line
		}
	}()
	return nil
}
