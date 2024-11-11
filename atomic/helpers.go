package main

import "strings"

// replaceStringInSlice simple helper to replace a given string in a slice of
// strings
func replaceStringInSlice(ss []string, replace string, with string) []string {
	result := []string{}
	for _, r := range ss {
		result = append(result, strings.Replace(
			r, replace, with, -1),
		)
	}

	return result
}
