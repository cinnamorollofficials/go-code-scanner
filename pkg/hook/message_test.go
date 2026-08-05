package hook

import "testing"

func TestValidateCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		pattern string
		max     int
		wantErr bool
	}{
		{name: "conventional", message: "feat(cli): add hook support\n", max: 72},
		{name: "breaking", message: "feat(api)!: remove legacy route\n", max: 72},
		{name: "comments ignored", message: "# template\n\nfix: repair parser\n", max: 72},
		{name: "custom pattern", message: "PROJ-42 describe change", pattern: `^PROJ-[0-9]+ .+`},
		{name: "empty", message: "# comments only\n", wantErr: true},
		{name: "invalid shape", message: "updated stuff", wantErr: true},
		{name: "too long", message: "feat: this subject is too long", max: 10, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCommitMessage(test.message, test.pattern, test.max)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateCommitMessage() error=%v wantErr=%t", err, test.wantErr)
			}
		})
	}
}
