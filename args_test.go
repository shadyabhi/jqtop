package main

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestParseArgs(t *testing.T) {
	origArgs := args
	os.Args = []string{"jtop", "-c", "-v", "--file", "./tailit", "--fields", "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");", "--filters", "contains(cquuc, \"media.licdn.com\"); contains(cquuc, \"mpr\");"}

	parseArgs()
	// Check default values
	if (args.Interval != 1) && (args.MaxResult != 10) {
		t.Errorf("Default argument was not set correctly")
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Errorf("Debug was not set!")
	}
	if args.Fields != "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");" {
		t.Errorf("args.Fields value is not as expected")
	}

	// Reset args
	args = origArgs
	// We know this was changed in the test
	logrus.SetLevel(logrus.InfoLevel)
}
