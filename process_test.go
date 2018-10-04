package main

import (
	"strconv"
	"testing"

	"github.com/tidwall/gjson"
)

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
	r := isMatching(`{"foo": "i am awesome"}`, []filter{})
	if !r {
		t.Errorf("No filters provided but line was filtered")
	}

	for i, tt := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := isMatching(tt.json, []filter{*tt.f})
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
			r := isMatchFilter(tt.json, *tt.f)
			if r != tt.expected {
				t.Errorf("Unexpected return from isMatchFilter. Input: %+v", tt)
			}

		})
	}
}

func TestProcessLine(t *testing.T) {
	args.Interval = 1

	line := []string{
		`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`,
	}

	simpleF := []string{"cquuc", "pssc", "cqhm", "cqhv"}

	complexFields["host_with_protocol"] = &complexField{
		NewField: "host_with_protocol",
		Fname:    "regex_capture",
		Args:     []string{"cquuc", "(.*?://.*?)/"},
	}

	processLine(line[0], simpleF, complexFields)

	if len(countersMap.counters) != 5 {
		t.Errorf("processLine didn't parse all fields")
	}
}

func BenchmarkProcessLine(b *testing.B) {
	args.Interval = 1

	line := []string{
		`{"cqtn": "23/Sep/2018:02:34:25 -0000", "cqhm": "GET", "cquuc": "http://coolhost.com:1234/admin", "cqhv": "HTTP/1.1", "pssc": "200"}`,
	}

	simpleF := []string{"cquuc", "pssc", "cqhm", "cqhv"}

	complexFields["host_with_protocol"] = &complexField{
		NewField: "host_with_protocol",
		Fname:    "regex_capture",
		Args:     []string{"cquuc", "(.*?://.*?)/"},
	}

	for n := 0; n < b.N; n++ {
		processLine(line[0], simpleF, complexFields)
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
