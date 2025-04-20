package utils

import (
	"slices"
)

func SliceSame(slice1, slice2 []any) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for _, elem := range slice1 {
		if !slices.Contains(slice2, elem) {
			return false
		}
	}
	for _, elem := range slice2 {
		if !slices.Contains(slice1, elem) {
			return false
		}
	}
	return true
}
