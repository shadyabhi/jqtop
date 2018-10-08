package jqtop

import (
	"strconv"
	"testing"

	"github.com/hpcloud/tail"
	"github.com/tidwall/gjson"
)

// clearCounters helps us to clear counters
// in tests
func clearCounters() {
	// Clear counters
	for k := range countersMap.counters {
		delete(countersMap.counters, k)
	}
}

func TestIsMatching(t *testing.T) {
	var testCases = []struct {
		json     string
		f        *filter
		expected bool
	}{
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "contains", Args: []string{"foo", "awesome"}},
			true,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "contains", Args: []string{"foo", "awesome"}},
			true,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "equals", Args: []string{"foo", "awesome"}},
			false,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Negate: false, Function: "contains", Args: []string{"foo", "awesome"}},
			true,
		},
	}
	r := isMatching(`{"foo": "i am awesome"}`, []filter{}, Fields{})
	if !r {
		t.Errorf("No filters provided but line was filtered")
	}

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := isMatching(tt.json, []filter{*tt.f}, Fields{})
			if r != tt.expected {
				t.Errorf("Unexpected return from isMatchFilter. Input: %+v", tt)
			}

		})
	}
}
func TestIsMatchFilter(t *testing.T) {
	var testCases = []struct {
		json     string
		f        *filter
		expected bool
	}{
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "contains", Args: []string{"foo", "awesome"}},
			true,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "contains", Args: []string{"not_found", "awesome"}},
			false,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "contains", Args: []string{"foo", "awesomeness"}},
			false,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "regex", Args: []string{"foo", "awe.*"}},
			true,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "regex", Args: []string{"foo", "*.*"}},
			false,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "regex", Args: []string{"foo", "AWE.*"}},
			false,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "equals", Args: []string{"foo", "i am awesome"}},
			true,
		},
		{
			`{"foo": "i am awesome"}`,
			&filter{Function: "equals_invalid_spelling", Args: []string{"foo", "i am awesome"}},
			false,
		},
	}
	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := isMatchFilter(tt.json, *tt.f, Fields{})
			if r != tt.expected {
				t.Errorf("Unexpected return from isMatchFilter. Input: %+v", tt)
			}

		})
	}
}

func TestProcessLine(t *testing.T) {
	args.Interval = 1
	clearCounters()

	line := []string{
		`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`,
	}

	simpleF := []string{"cquuc", "pssc", "cqhm", "cqhv"}
	derivedFields := make(map[string]*derivedField)
	derivedFields["host_with_protocol"] = &derivedField{
		NewField: "host_with_protocol",
		Fname:    "regex_capture",
		Args:     []string{"cquuc", "(.*?://.*?)/"},
	}
	allFields := Fields{
		SimpleFields:  simpleF,
		DerivedFields: derivedFields,
	}

	processLine(line[0], allFields)

	if len(countersMap.counters) != 5 {
		t.Errorf("processLine didn't parse all fields\n. Current counters: %+v", countersMap.counters)
	}
}

func TestProcessLines(t *testing.T) {
	// Set proper args and vars
	args.Interval = 1
	args.Fields = "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");"
	args.Filters = ""
	clearCounters()

	linesChan := make(chan *tail.Line)

	// Send lines to channel
	go func() {
		linesChan <- tail.NewLine(`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`)
		linesChan <- tail.NewLine(`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost2.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`)
		close(linesChan)
	}()

	// Reading from channel
	ProcessLines(linesChan)
	if len(countersMap.counters) != 2 {
		t.Errorf("Correct numbers of counters were not updated")
	}
	if countersMap.counters["cquuc"]["http://coolhost.com:1234/admin"].String() != "1" {
		t.Errorf("Wrong counter values")
	}
}

func BenchmarkProcessLine(b *testing.B) {
	args.Interval = 1

	line := []string{
		`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`,
	}

	simpleF := []string{"cquuc", "pssc", "cqhm", "cqhv"}
	derivedFields := make(map[string]*derivedField)
	derivedFields["host_with_protocol"] = &derivedField{
		NewField: "host_with_protocol",
		Fname:    "regex_capture",
		Args:     []string{"cquuc", "(.*?://.*?)/"},
	}
	allFields := Fields{
		SimpleFields:  simpleF,
		DerivedFields: derivedFields,
	}

	for n := 0; n < b.N; n++ {
		processLine(line[0], allFields)
	}
}

func BenchmarkGjson(b *testing.B) {

	line := `{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`

	b.Run("Individual fetch", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = gjson.Get(line, "pssc")
		}

	})
	b.Run("Individual fetches of multiple values", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = gjson.Get(line, "cquuc")
			_ = gjson.Get(line, "pssc")
			_ = gjson.Get(line, "cqhm")
			_ = gjson.Get(line, "cqhv")
		}

	})
	b.Run("GetMany: Many at once", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = gjson.GetMany(line, "cquuc", "pssc", "cqhm", "cqhv")
		}

	})
	/*
		goos: linux
		goarch: amd64
		PASS
		benchmark                                                    iter       time/iter
		---------                                                    ----       ---------
		BenchmarkGjson/Individual_fetch-12                        2000000    898.00 ns/op
		BenchmarkGjson/Individual_fetches_of_multiple_values-12    300000   3774.00 ns/op
		BenchmarkGjson/GetMany:_Many_at_once-12                    500000   4789.00 ns/op
		ok      _/home/arastogi/repos/jtop      6.416s
	*/
}
