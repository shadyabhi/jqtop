package jqtop

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestParseArgs(t *testing.T) {
	origArgs := Config
	os.Args = []string{"jtop", "-c", "-v", "--file", "./tailit", "--fields", "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");", "--filters", "contains(cquuc, \"media.licdn.com\"); contains(cquuc, \"mpr\");"}

	if err := ParseArgs(); err != nil {
		t.Errorf("Didn't expect error, got error: %s", err)
	}
	// Check default values
	if (Config.Interval != 1) && (Config.MaxResult != 10) {
		t.Errorf("Default argument was not set correctly")
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Errorf("Debug was not set!")
	}
	if Config.Fields != "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");" {
		t.Errorf("args.Fields value is not as expected")
	}

	// Reset args
	Config = origArgs
	// We know this was changed in the test
	logrus.SetLevel(logrus.InfoLevel)
}
