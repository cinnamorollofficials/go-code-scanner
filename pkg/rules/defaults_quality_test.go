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
		"trailing-whitespace": {
			positive: "const clean = true;  ",
			negative: "const clean = true;",
		},
		"mixed-indentation": {
			positive: " \treturn value",
			negative: "\treturn value",
		},
		"javascript-console-debug": {
			positive: "console.debug(result)",
			negative: "logger.DebugContext(ctx, result)",
		},
	})
}
