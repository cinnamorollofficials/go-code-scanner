package rules

import "testing"

func TestDefaultGovernanceRuleExamples(t *testing.T) {
	testRuleExamples(t, DefaultGovernance(), map[string]ruleExample{
		"privacy-pii-log": {
			positive: `log.Printf("customer email=%s", customer.Email)`,
			negative: `log.Printf("customer_id=%s", customer.ID)`,
		},
		"privacy-pii-url": {
			positive: `request.URL.Query().Set("email", customer.Email)`,
			negative: `request.URL.Query().Set("page", page)`,
		},
		"privacy-pii-fixture": {
			positive: `{"email": "person@example.com"}`,
			negative: `{"email": "${TEST_EMAIL}"}`,
		},
		"privacy-sensitive-response": {
			positive: `c.JSON(200, map[string]any{"ssn": customer.SSN})`,
			negative: `c.JSON(200, CustomerResponse{ID: customer.ID})`,
		},
	})
}
