package rules

import "testing"

func TestDefaultReliabilityRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultReliability(), map[string]ruleExample{
		"go-multipart-memory": {
			positive: `file, err := c.FormFile("upload")`,
			negative: `file, err := form.File["upload"]`,
		},
		"go-http-default-server": {
			positive: `log.Fatal(http.ListenAndServe(":8080", handler))`,
			negative: `log.Fatal(server.ListenAndServe())`,
		},
		"go-unbounded-request-read": {
			positive: `body, err := io.ReadAll(r.Body)`,
			negative: `body, err := io.ReadAll(io.LimitReader(r.Body, maxBody))`,
		},
	})
}
