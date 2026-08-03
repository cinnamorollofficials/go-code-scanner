package frontend

import (
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
)

type RuleDefinition struct {
	ID             string           `json:"id"`
	Domain         finding.Domain   `json:"domain"`
	Category       string           `json:"category"`
	Severity       finding.Severity `json:"severity"`
	Description    string           `json:"description"`
	Recommendation string           `json:"recommendation,omitempty"`
	Documentation  string           `json:"documentation,omitempty"`
	Framework      string           `json:"framework,omitempty"`
	Confidence     string           `json:"confidence,omitempty"`
	Sink           string           `json:"sink,omitempty"`
	Source         string           `json:"source,omitempty"`
	Tags           []string         `json:"tags,omitempty"`
}

func (rd RuleDefinition) ToRule() rules.Rule {
	return rules.Rule{
		ID:             rd.ID,
		Domain:         rd.Domain,
		Category:       rd.Category,
		Severity:       rd.Severity,
		Description:    rd.Description,
		Recommendation: rd.Recommendation,
		Documentation:  rd.Documentation,
		Tags:           rd.Tags,
	}
}

var Registry = map[string]RuleDefinition{
	"frontend/dom-injection": {
		ID:             "frontend/dom-injection",
		Domain:         finding.Security,
		Category:       "injection",
		Severity:       finding.High,
		Description:    "Potential DOM injection sink receiving un-sanitized dynamic input",
		Recommendation: "Sanitize dynamic input with DOMPurify or use safe DOM node assignment methods",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md#dom-injection",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "innerHTML",
		Source:         "location.search",
		Tags:           []string{"frontend", "xss", "dom"},
	},
	"frontend/unsafe-execution": {
		ID:             "frontend/unsafe-execution",
		Domain:         finding.Security,
		Category:       "execution",
		Severity:       finding.Critical,
		Description:    "Dynamic string code execution sink",
		Recommendation: "Avoid string-based eval or dynamic function constructors",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md#dynamic-code-execution",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "eval",
		Source:         "user-input",
		Tags:           []string{"frontend", "eval", "code-execution"},
	},
	"frontend/unsafe-messaging": {
		ID:             "frontend/unsafe-messaging",
		Domain:         finding.Security,
		Category:       "messaging",
		Severity:       finding.High,
		Description:    "Wildcard postMessage origin or message listener missing origin validation",
		Recommendation: "Specify an explicit target origin for postMessage and validate event.origin in message event listeners",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "postMessage",
		Source:         "event.data",
		Tags:           []string{"frontend", "postmessage", "origin"},
	},
	"frontend/client-credential-exposure": {
		ID:             "frontend/client-credential-exposure",
		Domain:         finding.Security,
		Category:       "credential-exposure",
		Severity:       finding.High,
		Description:    "Sensitive credential or secret key stored in client storage or public environment variable",
		Recommendation: "Keep private keys on server side; do not prefix secrets with NEXT_PUBLIC_ or VITE_",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md#exposure-of-client-secrets",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "localStorage",
		Source:         "env",
		Tags:           []string{"frontend", "secrets", "storage"},
	},
	"frontend/unsafe-navigation": {
		ID:             "frontend/unsafe-navigation",
		Domain:         finding.Hardening,
		Category:       "navigation",
		Severity:       finding.Medium,
		Description:    "Untrusted navigation target or insecure transport protocol",
		Recommendation: "Validate redirect URLs against an allowlist and enforce HTTPS endpoints",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "vanilla",
		Confidence:     "MEDIUM",
		Sink:           "location.href",
		Source:         "user-input",
		Tags:           []string{"frontend", "navigation", "open-redirect"},
	},
	"frontend/unsafe-transport": {
		ID:             "frontend/unsafe-transport",
		Domain:         finding.Hardening,
		Category:       "transport",
		Severity:       finding.Medium,
		Description:    "Insecure HTTP protocol used for non-localhost API endpoint",
		Recommendation: "Enforce HTTPS for non-development remote endpoint connections",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "http://",
		Source:         "url-literal",
		Tags:           []string{"frontend", "http", "transport"},
	},
	"frontend/sensitive-query-param": {
		ID:             "frontend/sensitive-query-param",
		Domain:         finding.Security,
		Category:       "privacy",
		Severity:       finding.Medium,
		Description:    "Sensitive credential placed in URL query parameter",
		Recommendation: "Pass sensitive credentials in HTTP request headers or request body, not URL query strings",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "vanilla",
		Confidence:     "HIGH",
		Sink:           "URLSearchParams",
		Source:         "query-param",
		Tags:           []string{"frontend", "url", "query-param"},
	},
	"frontend/telemetry-privacy-leak": {
		ID:             "frontend/telemetry-privacy-leak",
		Domain:         finding.Governance,
		Category:       "privacy",
		Severity:       finding.Low,
		Description:    "Potential PII or sensitive payload logged to client console or analytics",
		Recommendation: "Sanitize log payloads to include safe metadata identifiers only",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md#privacy-telemetry-leaks",
		Framework:      "vanilla",
		Confidence:     "MEDIUM",
		Sink:           "console.log",
		Source:         "user-object",
		Tags:           []string{"frontend", "privacy", "pii"},
	},
	"frontend/react-dangerously-set-inner-html": {
		ID:             "frontend/react-dangerously-set-inner-html",
		Domain:         finding.Security,
		Category:       "injection",
		Severity:       finding.High,
		Description:    "React component using dangerouslySetInnerHTML with un-sanitized dynamic markup",
		Recommendation: "Sanitize HTML using DOMPurify before setting dangerouslySetInnerHTML",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "react",
		Confidence:     "HIGH",
		Sink:           "dangerouslySetInnerHTML",
		Source:         "props",
		Tags:           []string{"frontend", "react", "xss"},
	},
	"frontend/next-server-module-in-client": {
		ID:             "frontend/next-server-module-in-client",
		Domain:         finding.Hardening,
		Category:       "boundary",
		Severity:       finding.High,
		Description:    "Client component importing Node/server-only module",
		Recommendation: "Move server-only logic to server components or API routes",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "next",
		Confidence:     "HIGH",
		Sink:           "import",
		Source:         "server-only-module",
		Tags:           []string{"frontend", "next", "architecture", "server-only"},
	},
	"frontend/next-private-env-in-client": {
		ID:             "frontend/next-private-env-in-client",
		Domain:         finding.Security,
		Category:       "credential-exposure",
		Severity:       finding.High,
		Description:    "Client module attempting to read non-public server environment variable",
		Recommendation: "Only access public NEXT_PUBLIC_ variables on the client side",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "next",
		Confidence:     "HIGH",
		Sink:           "process.env",
		Source:         "private-env",
		Tags:           []string{"frontend", "next", "env", "secrets"},
	},
	"frontend/vue-v-html": {
		ID:             "frontend/vue-v-html",
		Domain:         finding.Security,
		Category:       "injection",
		Severity:       finding.High,
		Description:    "Vue directive v-html bound to un-sanitized expression",
		Recommendation: "Sanitize dynamic HTML content before rendering via v-html",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "vue",
		Confidence:     "HIGH",
		Sink:           "v-html",
		Source:         "data",
		Tags:           []string{"frontend", "vue", "xss"},
	},
	"frontend/nuxt-private-runtime-config": {
		ID:             "frontend/nuxt-private-runtime-config",
		Domain:         finding.Security,
		Category:       "credential-exposure",
		Severity:       finding.High,
		Description:    "Client code accessing private Nuxt runtimeConfig field",
		Recommendation: "Move secret configuration fields to public runtimeConfig object or keep logic on server side",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "nuxt",
		Confidence:     "HIGH",
		Sink:           "useRuntimeConfig",
		Source:         "private-runtime-config",
		Tags:           []string{"frontend", "nuxt", "config", "secrets"},
	},
	"frontend/nuxt-server-import-in-client": {
		ID:             "frontend/nuxt-server-import-in-client",
		Domain:         finding.Hardening,
		Category:       "boundary",
		Severity:       finding.High,
		Description:    "Client code importing Nuxt server directory or server util module",
		Recommendation: "Do not import server-only modules in client components; invoke server endpoints via $fetch",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "nuxt",
		Confidence:     "HIGH",
		Sink:           "import",
		Source:         "server-dir",
		Tags:           []string{"frontend", "nuxt", "architecture", "server-only"},
	},
	"frontend/svelte-html": {
		ID:             "frontend/svelte-html",
		Domain:         finding.Security,
		Category:       "injection",
		Severity:       finding.High,
		Description:    "Svelte {@html} tag rendering un-sanitized dynamic expression",
		Recommendation: "Sanitize dynamic markup before rendering with {@html}",
		Documentation:  "https://github.com/cinnamorollofficials/go-code-scanner/blob/main/docs/frontend-scanning-contract.md",
		Framework:      "svelte",
		Confidence:     "HIGH",
		Sink:           "{@html}",
		Source:         "props",
		Tags:           []string{"frontend", "svelte", "xss"},
	},
}

func LookupRule(id string) (RuleDefinition, bool) {
	rule, ok := Registry[id]
	return rule, ok
}

func RuleDefinitions() []rules.Rule {
	list := make([]rules.Rule, 0, len(Registry))
	for _, def := range Registry {
		list = append(list, def.ToRule())
	}
	return list
}
