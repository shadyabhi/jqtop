package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestTailThis(t *testing.T) {
	tmpfile, err := ioutil.TempFile("./", "tailfile")
	fLines, err := tailThis(tmpfile.Name())
	if err != nil {
		t.Errorf("Error reading the file, not expected")
	}

	go func() {
		for {
			tmpfile.Write([]byte("foo\n"))
		}
	}()

	totalLines := 0

	for range fLines.Lines {
		totalLines = totalLines + 1
		if totalLines == 2 {
			break
		}
	}
	if totalLines != 2 {
		t.Errorf("Expected two lines to be read, but only read %d lines", totalLines)
	}

	if err = os.Remove(tmpfile.Name()); err != nil {
		t.Errorf("There was error deleting the file: %s", err)
	}

}
