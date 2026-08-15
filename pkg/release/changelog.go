package release

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var changelogVersion = regexp.MustCompile(`^## \[([0-9]+)\.([0-9]+)\.([0-9]+)\] - ([0-9]{4}-[0-9]{2}-[0-9]{2})$`)

func ValidateChangelog(data []byte) error {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.HasPrefix(text, "---\n") {
		if endIdx := strings.Index(text[4:], "\n---\n"); endIdx != -1 {
			text = strings.TrimLeft(text[4+endIdx+5:], " \t\r\n")
		}
	}
	if !strings.HasPrefix(text, "# Changelog\n") {
		return fmt.Errorf("changelog must start with # Changelog")
	}
	if !strings.Contains(text, "\n## [Unreleased]\n") {
		return fmt.Errorf("changelog must contain an Unreleased section")
	}
	previous := [3]int{int(^uint(0) >> 1), 0, 0}
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "## [") || line == "## [Unreleased]" {
			continue
		}
		matches := changelogVersion.FindStringSubmatch(line)
		if matches == nil {
			return fmt.Errorf("invalid changelog release heading %q", line)
		}
		current := [3]int{}
		for index := range 3 {
			current[index], _ = strconv.Atoi(matches[index+1])
		}
		if compareVersion(current, previous) >= 0 {
			return fmt.Errorf("changelog releases must be newest first")
		}
		previous = current
	}
	return nil
}

func compareVersion(left, right [3]int) int {
	for index := range 3 {
		if left[index] < right[index] {
			return -1
		}
		if left[index] > right[index] {
			return 1
		}
	}
	return 0
}
