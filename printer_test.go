package jqtop

import (
	"reflect"
	"testing"

	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

func Test_getSortedCounters(t *testing.T) {
	// Setup
	// Fields
	fieldsStr := "cquuc"
	allFields, err := extractFields(fieldsStr)
	if err != nil {
		logrus.Fatalf("Error parsing fields, existing")
	}

	fMap := getFieldIndexMap(allFields.FieldsInOrder)
	// Holds last value (avoid timers)
	// Array of map[fieldname](counterValue)
	var lastCounters []map[string]int64
	initLastCounters(&lastCounters, allFields)
	countersSlice.counters = initFieldCounters(allFields)

	r := metrics.NewRegistry()
	c1 := r.GetOrRegister("c1", metrics.NewCounter).(*metrics.StandardCounter)
	c1.Inc(30)
	c2 := r.GetOrRegister("c2", metrics.NewCounter).(*metrics.StandardCounter)
	c2.Inc(20)
	c3 := r.GetOrRegister("c3", metrics.NewCounter).(*metrics.StandardCounter)
	c3.Inc(24)
	c4 := r.GetOrRegister("c4", metrics.NewCounter).(*metrics.StandardCounter)
	c4.Inc(26)
	countersSlice.counters[0] = r

	ss := getSortedCounters(&fMap, "cquuc", &lastCounters)
	expected := []sortedCounters{
		{"c1", 30, 30},
		{"c4", 26, 26},
		{"c3", 24, 24},
		{"c2", 20, 20},
	}
	if !reflect.DeepEqual(ss, expected) {
		t.Errorf("Sorted counters returned have wrong order, returned: %#v, expected: %#v", ss, expected)
	}
}
