package jqtop

import (
	"testing"
)

func TestTailThis(t *testing.T) {
	tmpFile, _ := getFileToTail()
	defer deleteTmpFile(tmpFile)

	fLines, err := TailFile(tmpFile.Name())
	if err != nil {
		t.Errorf("Error reading the file, not expected")
	}

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

	// File doesn't exist
	_, err = TailFile("foobarnotexist")
	if err == nil {
		t.Errorf("Expected error as file doesn't exist, got no error")
	}
}
