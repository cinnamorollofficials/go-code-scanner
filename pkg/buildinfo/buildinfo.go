package buildinfo

import "strings"

var (
	Version = "0.1.0-dev"
	Commit  = ""
	Date    = ""
)

// String returns stable release metadata populated through -ldflags. Local
// development builds retain the backward-compatible version-only output.
func String() string {
	parts := []string{Version}
	if strings.TrimSpace(Commit) != "" {
		parts = append(parts, "commit="+Commit)
	}
	if strings.TrimSpace(Date) != "" {
		parts = append(parts, "built="+Date)
	}
	return strings.Join(parts, " ")
}
