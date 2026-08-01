package finding

import "testing"

func TestParseDomain(t *testing.T) {
	domain, err := ParseDomain(" Supply_Chain ")
	if err != nil {
		t.Fatal(err)
	}
	if domain != SupplyChain {
		t.Fatalf("expected %q, got %q", SupplyChain, domain)
	}
}

func TestParseDomainRejectsUnknownValue(t *testing.T) {
	if _, err := ParseDomain("performance"); err == nil {
		t.Fatal("expected invalid domain error")
	}
}

func TestAllDomainsAreValid(t *testing.T) {
	domains := []Domain{Quality, Reliability, Hardening, Security, SupplyChain, Governance}
	for _, domain := range domains {
		if !domain.Valid() {
			t.Fatalf("expected domain %q to be valid", domain)
		}
	}
}
