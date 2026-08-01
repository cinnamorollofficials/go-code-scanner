package rules

import "testing"

func TestDefaultQualityRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultQuality(), map[string]ruleExample{
		"merge-conflict-marker": {
			positive: "<<<<<<< HEAD",
			negative: "// explain what a merge conflict marker looks like",
		},
		"javascript-debugger": {
			positive: "  debugger;  ",
			negative: "const debuggerEnabled = false;",
		},
	})
}
