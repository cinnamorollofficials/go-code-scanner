package hook

import (
	"fmt"
	"regexp"
	"strings"
)

const ConventionalCommitPattern = `^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9][a-z0-9._/-]*\))?!?: .+`

func ValidateCommitMessage(content, pattern string, maxSubjectLength int) error {
	subject := commitSubject(content)
	if subject == "" {
		return fmt.Errorf("commit message subject is required")
	}
	if maxSubjectLength > 0 && len([]rune(subject)) > maxSubjectLength {
		return fmt.Errorf("commit message subject exceeds %d characters", maxSubjectLength)
	}
	if pattern == "" {
		pattern = ConventionalCommitPattern
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("compile commit message pattern: %w", err)
	}
	if !compiled.MatchString(subject) {
		return fmt.Errorf("commit message subject does not match required pattern")
	}
	return nil
}

func commitSubject(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}
