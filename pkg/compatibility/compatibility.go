package compatibility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/baseline"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/cache"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/hook"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/release"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/suppression"
)

type Change struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

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

func Decode(data []byte) (Contract, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var contract Contract
	if err := decoder.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("decode compatibility contract: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return Contract{}, fmt.Errorf("decode compatibility contract: %w", err)
	}
	return contract, nil
}

func Compare(previous, current Contract) []Change {
	changes := make([]Change, 0)
	add := func(field string, from, to any) {
		if from != to {
			changes = append(changes, Change{Field: field, From: fmt.Sprint(from), To: fmt.Sprint(to)})
		}
	}
	add("config_schema", previous.ConfigSchema, current.ConfigSchema)
	add("report_schema", previous.ReportSchema, current.ReportSchema)
	add("rule_schema", previous.RuleSchema, current.RuleSchema)
	add("suppression_schema", previous.SuppressionSchema, current.SuppressionSchema)
	add("baseline_schema", previous.BaselineSchema, current.BaselineSchema)
	add("fingerprint_version", previous.FingerprintVersion, current.FingerprintVersion)
	add("hook_marker_version", previous.HookMarkerVersion, current.HookMarkerVersion)
	add("cache_key_version", previous.CacheKeyVersion, current.CacheKeyVersion)
	add("provenance_schema", previous.ProvenanceSchema, current.ProvenanceSchema)
	return changes
}
