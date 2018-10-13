package jqtop

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
)

func getStdinStream() *os.File {
	stdin := os.Stdin

	go func() {
		stdin.Write([]byte("{\"fname\": \"foobar\"}\n"))
	}()

	return stdin
}

func readFewLines(linesChan chan *tail.Line, n int) (totalLines int) {
	for range linesChan {
		totalLines = totalLines + 1
		if totalLines == n {
			break
		}
	}
	return totalLines
}

func deleteTmpFile(tmpFile *os.File) (err error) {
	if err = os.Remove(tmpFile.Name()); err != nil {
		return errors.Wrap(err, "There was error deleting the file")
	}
	return nil
}

func TestTailStdin(t *testing.T) {

	stdin := getStdinStream()
	linesChan := make(chan *tail.Line)
	tailStdin(linesChan, stdin)

	totalLines := readFewLines(linesChan, 2)

	if totalLines != 2 {
		t.Errorf("Expected two lines to be read, but only read %d lines", totalLines)
	}
}

// getFileToTail creates a temp file and starts writing
// json to it
// NOTE: Caller must call deleteTmpFile function to clear
// the file
func getFileToTail() (tmpFile *os.File, err error) {
	tmpFile, err = ioutil.TempFile("./", "tailfile")

	if err != nil {
		return nil, errors.Wrap(err, "Error getting file to tail: %s")
	}
	go func() {
		for {
			tmpFile.Write([]byte("{\"name\": \"foobar\"}\n"))
		}
	}()
	return tmpFile, nil
}

func TestTailF(t *testing.T) {
	tmpfile, err := getFileToTail()
	defer deleteTmpFile(tmpfile)
	if err != nil {
		t.Errorf("Error getting file to tail: %s", err)
	}

	linesChan := make(chan *tail.Line)
	if err := tailF(linesChan, tmpfile.Name()); err != nil {
		t.Errorf("Error reading the file, not expected")
	}

	totalLines := readFewLines(linesChan, 2)
	if totalLines != 2 {
		t.Errorf("Expected two lines to be read, but only read %d lines", totalLines)
	}

	if err := tailF(linesChan, "file_doesnt_exist"); err == nil {
		t.Errorf("File doesn't exist, expected error")
	}

}

func TestSetLinesChan(t *testing.T) {
	linesChan := make(chan *tail.Line)

	// Stdin
	setLinesChan("", linesChan)
	totalLines := readFewLines(linesChan, 2)
	if totalLines != 2 {
		t.Errorf("Expected two lines to be read, but only read %d lines", totalLines)
	}

	//File
	linesChan2 := make(chan *tail.Line)
	tmpFile, err := getFileToTail()
	defer deleteTmpFile(tmpFile)
	if err != nil {
		t.Errorf("Error getting file to tail: %s", err)
	}
	setLinesChan(tmpFile.Name(), linesChan2)
	totalLines = readFewLines(linesChan2, 2)

	if totalLines != 2 {
		t.Errorf("Expected two lines to be read, but only read %d lines", totalLines)
	}

	// File doesnt exist
	if err := setLinesChan("file_doesnt_exist", linesChan2); err == nil {
		t.Errorf("Expected error as file doesn't exist")
	}
}

func TestStart(t *testing.T) {
	os.Args = []string{"jtop", "--fields", "fname", "-i", "0.01"}

	// Start json stream on stdin
	getStdinStream()

	buf := make([]byte, 1000)
	outStream := bytes.NewBuffer(buf)

	go Start(outStream)

	time.Sleep(15 * time.Millisecond)

	// Better than infinite for loop
	timeout := time.After(1 * time.Second)
	tick := time.Tick(5 * time.Millisecond)
	select {
	case <-timeout:
		t.Errorf("No counters were updated after timeout")

	case <-tick:
		if len(countersMap.counters) > 0 {
			break
		}
	}

}
