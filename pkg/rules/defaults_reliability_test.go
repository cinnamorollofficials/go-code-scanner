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
		"go-discarded-error": {
			positive: `_ = response.Body.Close()`,
			negative: `if err := response.Body.Close(); err != nil { return err }`,
		},
		"go-process-termination": {
			positive: `log.Fatalf("serve: %v", err)`,
			negative: `return fmt.Errorf("serve: %w", err)`,
		},
		"go-http-client-without-timeout": {
			positive: `client := &http.Client{}`,
			negative: `client := &http.Client{Timeout: 10 * time.Second}`,
		},
	})
}
