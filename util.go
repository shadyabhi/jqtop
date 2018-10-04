package main

import "reflect"

func getFieldValues(s interface{}) []string {
	v := reflect.ValueOf(s)
	values := make([]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface().(string)
	}
	return values
}

// sliceContains checks if a slice contains "s"
func sliceContains(slice []string, s string) bool {
	for _, op := range slice {
		if s == op {
			return true
		}
	}
	return false
}
