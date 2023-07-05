package utils

import (
	"strings"
)

func ShouldParse(group string, ignoregroup, includegroup []string) bool {
	// Core groups should never be ignored
	if group == "" || !strings.Contains(group, ".") {
		return true
	}
	// If there is an exact group to be ignored, ignore it.
	for _, ignore := range ignoregroup {
		if group == ignore {
			return false
		}
	}

	// If this is a group that should be included by prefix, then include it
	for _, include := range includegroup {
		if strings.HasSuffix(group, include) {
			return true
		}
	}

	// If includegroup exists, and the item is not inside the list
	return len(includegroup) == 0
}
