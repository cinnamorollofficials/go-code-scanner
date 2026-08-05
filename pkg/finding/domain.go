package finding

import (
	"fmt"
	"strings"
)

// Domain identifies the policy area responsible for a finding.
type Domain string

const (
	Quality     Domain = "quality"
	Reliability Domain = "reliability"
	Hardening   Domain = "hardening"
	Security    Domain = "security"
	SupplyChain Domain = "supply_chain"
	Governance  Domain = "governance"
)

var validDomains = map[Domain]struct{}{
	Quality:     {},
	Reliability: {},
	Hardening:   {},
	Security:    {},
	SupplyChain: {},
	Governance:  {},
}

func ParseDomain(value string) (Domain, error) {
	domain := Domain(strings.ToLower(strings.TrimSpace(value)))
	if !domain.Valid() {
		return "", fmt.Errorf("invalid finding domain %q", value)
	}
	return domain, nil
}

func (d Domain) Valid() bool {
	_, ok := validDomains[d]
	return ok
}
