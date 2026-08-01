package compatibility

import (
	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/baseline"
	"github.com/cinnamorollofficials/go-code-scanner/cache"
	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/hook"
	"github.com/cinnamorollofficials/go-code-scanner/release"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/suppression"
)

type Contract struct {
	ConfigSchema       int    `json:"config_schema"`
	ReportSchema       string `json:"report_schema"`
	RuleSchema         int    `json:"rule_schema"`
	SuppressionSchema  int    `json:"suppression_schema"`
	BaselineSchema     int    `json:"baseline_schema"`
	FingerprintVersion string `json:"fingerprint_version"`
	HookMarkerVersion  string `json:"hook_marker_version"`
	CacheKeyVersion    string `json:"cache_key_version"`
	ProvenanceSchema   string `json:"provenance_schema"`
}

func Current() Contract {
	return Contract{
		ConfigSchema: config.SchemaVersion, ReportSchema: securityreview.SchemaVersion,
		RuleSchema: rules.SchemaVersion, SuppressionSchema: suppression.SchemaVersion,
		BaselineSchema: baseline.Version, FingerprintVersion: securityreview.FingerprintVersion,
		HookMarkerVersion: hook.MarkerVersion, CacheKeyVersion: cache.KeyVersion,
		ProvenanceSchema: release.ProvenanceSchema,
	}
}
