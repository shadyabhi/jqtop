package jqtop

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/hpcloud/tail"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/tidwall/gjson"
)

// clearCounters helps us to clear counters
// in tests
func clearCounters() {
	// Clear counters

	for _, r := range countersSlice.counters {
		r.Each(func(name string, i interface{}) {
			c := i.(*metrics.StandardCounter)
			c.Clear()
		})

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
	Config.Interval = 1
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
		FieldsInOrder: []string{"cquuc", "pssc", "cqhm", "cqhv", "host_with_protocol"},
	}
	// This is initialized elsewhere, needed for test only
	countersSlice.counters = initFieldCounters(allFields)

	processLine(line[0], allFields)

	if len(countersSlice.counters) != 5 {
		t.Errorf("processLine didn't parse all fields\n. Current counters: %+v", countersSlice.counters)
	}
}

func TestProcessLines(t *testing.T) {
	// Set proper args and vars
	Config.Interval = 1
	Config.Fields = "cquuc; host_with_protocol = regex_capture(cquuc, \"(.*?://.*?)/\");"
	Config.Filters = ""
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
	if len(countersSlice.counters) != 2 {
		t.Errorf("Correct numbers of counters were not updated")
	}
	c := countersSlice.counters[0].GetOrRegister("http://coolhost.com:1234/admin", metrics.NewCounter).(*metrics.StandardCounter)
	if c.Count() != int64(1) {
		t.Errorf("Wrong counter values")
	}
}

func BenchmarkProcessLine(b *testing.B) {
	Config.Interval = 1

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

func Test_getFieldIndex(t *testing.T) {
	type args struct {
		allFields Fields
		fName     string
	}
	tests := []struct {
		name      string
		args      args
		wantIndex int
	}{
		{"valid field at index 0", args{Fields{FieldsInOrder: []string{"cquuc", "domain"}}, "cquuc"}, 0},
		{"valid field at index 1", args{Fields{FieldsInOrder: []string{"cquuc", "domain"}}, "domain"}, 1},
		{"invalid field", args{Fields{FieldsInOrder: []string{"cquuc", "domain"}}, "non_existent"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIndex := getFieldIndex(tt.args.allFields, tt.args.fName); gotIndex != tt.wantIndex {
				t.Errorf("getIndexField() = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}

func Test_initFieldCounters(t *testing.T) {
	type args struct {
		allFields Fields
	}
	tests := []struct {
		name                 string
		args                 args
		wantRegistrySliceLen int
	}{
		{"valid", args{Fields{FieldsInOrder: []string{"cquuc", "domain"}}}, 2},
		{"no fields", args{Fields{FieldsInOrder: []string{}}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRegistrySlice := initFieldCounters(tt.args.allFields); !reflect.DeepEqual(gotRegistrySlice, tt.wantRegistrySliceLen) {
				if len(gotRegistrySlice) != tt.wantRegistrySliceLen {
					t.Errorf("initCounters() = %v, want %v", len(gotRegistrySlice), tt.wantRegistrySliceLen)
				}
			}
		})
	}
}
