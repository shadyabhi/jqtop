package jqtop

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/hpcloud/tail"
	"github.com/paulbellamy/ratecounter"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	metrics "github.com/rcrowley/go-metrics"
)

// countersSlice stores registries (go-metrics) for various fields
// We use a slice so we can find the registry without introducing locks (eg, using map)
var countersSlice struct {
	counters []metrics.Registry
	sync.RWMutex
}

var parseErrors *ratecounter.RateCounter

// ProcessLines reads from "lines" channel
// and processes them
func ProcessLines(lines chan *tail.Line) {
	// Init stuff
	// Filters
	filters, err := parseFilters(Config.Filters)
	if err != nil {
		logrus.Fatalf("Error parsing filters, existing")
	}

	// Fields
	allFields, err := extractFields(Config.Fields)
	if err != nil {
		logrus.Fatalf("Error parsing fields, existing")
	}

	// Counters
	countersSlice.Lock()
	countersSlice.counters = initFieldCounters(allFields)
	countersSlice.Unlock()

	// Go Routines
	var wg sync.WaitGroup
	for i := 0; i < Config.ParallelProc; i++ {
		wg.Add(1)
		go startProcessLines(lines, &allFields, filters, &wg)
	}
	wg.Wait()
}

func startProcessLines(lines chan *tail.Line, allFields *Fields, filters []filter, wg *sync.WaitGroup) {
	defer wg.Done()
	logrus.Debugf("Filters to apply: %+v", filters)

	for line := range lines {
		logrus.Debugf("processLines: Processing line: %s", line)
		if isMatching(line.Text, filters, allFields) {
			processLine(line.Text, allFields)
		} else {
			logrus.Debugf("processLines: Line was filtered: %s", line)
		}
	}
}

// isMatching filters line based on []filters provided
func isMatching(line string, filters []filter, allFields *Fields) bool {
	if len(filters) == 0 {
		return true
	}

	results := make([]bool, 0)

	for _, f := range filters {
		isMatch := isMatchFilter(line, f, allFields)

		if f.Negate {
			isMatch = !isMatch
		}

		results = append(results, isMatch)

		logrus.Debugf("isMatching: Filter passed: %+v, line: %s, isMatch: %t", f, line, isMatch)
	}

	// Any match, filter the line
	for _, r := range results {
		if !r {
			return false
		}
	}

	// Looks like all are true
	return true
}

// getAnyValue gets value of a field (simple or complex)
func getAnyValue(fieldName string, line string, allFields *Fields) (contents string, err error) {
	contents, exists := getValue(line, fieldName)
	if !exists {
		// Let's try complex fields
		contents, err = getComplexFieldValue(fieldName, line, allFields.DerivedFields)
		if err != nil {
			parseErrors.Incr(1)
		}
		return contents, nil
	}
	return contents, nil
}

// isMatchFilter decides if a line should be filtered or not
// based on the "filte" provided
func isMatchFilter(line string, f filter, allFields *Fields) bool {
	// Get value of field
	contents, err := getAnyValue(f.Args[0], line, allFields)
	if err != nil {
		parseErrors.Incr(1)
		return false
	}

	switch op := f.Function; op {
	case filterFunctions.Contains:
		return strings.Contains(contents, f.Args[1])
	case filterFunctions.Regex:
		match, err := regexp.MatchString(f.Args[1], contents)
		if err != nil {
			parseErrors.Incr(1)
		} else {
			return match
		}
	case filterFunctions.Equal:
		return f.Args[1] == contents
	default:
		parseErrors.Incr(1)
	}
	return false
}

// processLine process various kind of lines
func processLine(line string, allFields *Fields) {
	// Simple fields
	processSimpleFields(line, allFields)

	// Complex fields
	if err := processComplexFields(line, allFields); err != nil {
		parseErrors.Incr(1)
	}
}

// processSimpleFields works on fields that
// don't need modification
func processSimpleFields(line string, allFields *Fields) {
	fields := allFields.SimpleFields
	values := make(map[string]string)
	for _, f := range fields {
		value, exists := getValue(line, f)
		if !exists {
			parseErrors.Incr(1)
		}
		values[f] = value
	}
	processValues(values, allFields)
}

func getComplexFieldValue(fieldName string, line string, fields map[string]*derivedField) (string, error) {
	cField, ok := fields[fieldName]
	if !ok {
		return "", fmt.Errorf("Couldn't get value for %s", cField)
	}
	origValue, exists := getValue(line, fields[fieldName].Args[0])
	if !exists {
		return "", fmt.Errorf("Couldn't get value for %s", cField)
	}
	value, err := deriveValue(fields[fieldName], origValue)
	if err != nil {
		return "", fmt.Errorf("Couldn't derive value %s", cField)
	}
	return value, nil
}

// processComplexFields parses fields whose values
// are derived from other fields
func processComplexFields(line string, allFields *Fields) error {
	fields := allFields.DerivedFields

	values := make(map[string]string)
	for k, v := range fields {
		value, err := getComplexFieldValue(k, line, fields)
		if err != nil {
			return err
		}
		values[v.NewField] = value
	}
	processValues(values, allFields)

	return nil
}

// getValue parses field "f" from json "s"
func getValue(s string, f string) (string, bool) {
	result := gjson.Get(s, f)
	return result.String(), result.Exists()
}

// deriveValue uses regex "regex" and returns the match, else error
func deriveValue(f *derivedField, origValue string) (string, error) {
	// Currently, only supporting regex_capture
	r, err := regexp.Compile(f.Args[1])
	if err != nil {
		logrus.Fatalf("Invalid value of regex provided in fields: %s, error: %s", f.Args[0], err)
	}
	res := r.FindStringSubmatch(origValue)
	if len(res) < 2 {
		return "", errors.New("Not Found")
	}
	return res[1], nil
}

// processValues reads values and updates counters
func processValues(values map[string]string, allFields *Fields) {
	for field, value := range values {
		// Increment new counter
		// These counters keep getting added, for now, we don't clear them
		countersSlice.counters[allFields.FieldsIndexMap[field]].GetOrRegister(
			value, metrics.NewCounter).(*metrics.StandardCounter).Inc(1)

	}
}

// getFieldIndex gets the index of Field
func getFieldIndex(allFields Fields, fName string) (index int) {
	for i, field := range allFields.FieldsInOrder {
		if field == fName {
			return i
		}
	}
	return -1
}

// initFieldCounters gets slice of registry of counters for all fields
func initFieldCounters(allFields Fields) (registrySlice []metrics.Registry) {
	cSlice := make([]metrics.Registry, len(allFields.FieldsInOrder))
	for i := range allFields.FieldsInOrder {
		cSlice[i] = metrics.NewRegistry()
	}
	return cSlice
}
