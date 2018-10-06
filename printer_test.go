package main

import (
	"testing"
)

func TestGetSortedCounters(t *testing.T) {
	ensureCounter("key", "field_value")
	countersMap.counters["key"]["field_value"].Incr(1)

	ensureCounter("key", "field_value2")
	countersMap.counters["key"]["field_value2"].Incr(1)
	countersMap.counters["key"]["field_value2"].Incr(1)

	ensureCounter("key", "field_value3")
	countersMap.counters["key"]["field_value3"].Incr(1)
	countersMap.counters["key"]["field_value3"].Incr(1)
	countersMap.counters["key"]["field_value3"].Incr(1)

	sorted := getSortedCounters(countersMap.counters["key"])

	// field_value3 has the highest count, it should be first
	if sorted[0].Key != "field_value3" {
		t.Errorf("Expected field_value3 to be the first element, got %s", sorted[0].Key)
	}

}
