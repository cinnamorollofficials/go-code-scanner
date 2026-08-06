package pattern

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/scanner"
)

type Scanner struct {
	genericRules  []rules.Compiled
	rulesBySuffix map[string][]rules.Compiled
	workers       int
	limits        Limits
	headers       []compiledHeader
}

type Limits struct {
	MaxFileBytes         int64
	MaxLineBytes         int
	QualityMaxFileBytes  int64
	QualityMaxLineLength int
	DependencyAllowlist  []string
	DependencyDenylist   []string
	LicenseAllowlist     []string
	LicenseDenylist      []string
	RequiredFiles        []string
	RequiredHeaders      []HeaderPolicy
	OwnershipFile        string
	OwnershipRules       []OwnershipPolicy
}

type OwnershipPolicy struct {
	Path     string
	Owners   []string
	Severity finding.Severity
}

type HeaderPolicy struct {
	ID             string
	Paths          []string
	Pattern        string
	MaxLines       int
	Severity       finding.Severity
	Description    string
	Recommendation string
}

type compiledHeader struct {
	HeaderPolicy
	expression *regexp.Regexp
}

func New(compiled []rules.Compiled, workers int, configured ...Limits) *Scanner {
	limits := Limits{MaxFileBytes: 2 * 1024 * 1024, MaxLineBytes: 1024 * 1024}
	if len(configured) > 0 {
		limits = configured[0]
	}
	s := &Scanner{rulesBySuffix: make(map[string][]rules.Compiled), workers: max(workers, 1), limits: limits}
	for _, header := range limits.RequiredHeaders {
		expression, err := regexp.Compile(header.Pattern)
		if err == nil {
			s.headers = append(s.headers, compiledHeader{HeaderPolicy: header, expression: expression})
		}
	}
	for _, rule := range compiled {
		if len(rule.Extensions) == 0 {
			s.genericRules = append(s.genericRules, rule)
			continue
		}
		for _, suffix := range rule.Extensions {
			suffix = strings.ToLower(suffix)
			s.rulesBySuffix[suffix] = append(s.rulesBySuffix[suffix], rule)
		}
	}
	return s
}

func (s *Scanner) ID() string { return "pattern" }

func (s *Scanner) Describe() scanner.Descriptor {
	return scanner.Descriptor{
		Capabilities:   []string{"built-in-rules", "line-patterns", "redaction"},
		SupportedModes: []string{"full", "changed", "staged"},
	}
}

func (s *Scanner) Scan(ctx context.Context, request scanner.Request) scanner.Result {
	started := time.Now()
	result := scanner.Result{State: finding.ScannerClean}
	fileFindings, fileErrors := s.scanFilePolicies(ctx, request)
	result.Findings = append(result.Findings, fileFindings...)
	for _, err := range fileErrors {
		result.Message = appendMessage(result.Message, err.Error())
	}
	if len(request.Sources) == 0 {
		if result.Message != "" {
			result.State, result.Failure = finding.ScannerPartial, scanner.FailurePartial
		} else if len(result.Findings) > 0 {
			result.State = finding.ScannerFindings
		}
		result.Duration = time.Since(started)
		return result
	}
	jobs := make(chan scanner.Source)
	outcomes := make(chan outcome, len(request.Sources))
	workerCount := min(s.workers, len(request.Sources))
	for range workerCount {
		go func() {
			for source := range jobs {
				items, err := s.scanSource(ctx, source, request.Root)
				outcomes <- outcome{findings: items, err: err}
			}
		}()
	}
	go feedSources(ctx, request.Sources, jobs)
	for completed := 0; completed < len(request.Sources); completed++ {
		select {
		case item := <-outcomes:
			result.Findings = append(result.Findings, item.findings...)
			if item.err != nil {
				result.Message = appendMessage(result.Message, item.err.Error())
			}
		case <-ctx.Done():
			result.State, result.Message, result.Failure = finding.ScannerFailed, ctx.Err().Error(), scanner.FailureCanceled
			result.Duration = time.Since(started)
			return result
		}
	}
	if result.Message != "" {
		result.State = finding.ScannerPartial
		result.Failure = scanner.FailurePartial
	} else if len(result.Findings) > 0 {
		result.State = finding.ScannerFindings
	}
	result.Duration = time.Since(started)
	return result
}

func (s *Scanner) scanFilePolicies(ctx context.Context, request scanner.Request) ([]finding.Finding, []error) {
	var findings []finding.Finding
	var failures []error
	knownFiles := make(map[string]struct{}, len(request.Files))
	for _, source := range request.Files {
		relative, err := filepath.Rel(request.Root, source.Path)
		if err == nil {
			knownFiles[strings.ToLower(filepath.ToSlash(relative))] = struct{}{}
		}
	}
	repositoryFiles := make(map[string]struct{}, len(request.RepositoryFiles))
	for _, source := range request.RepositoryFiles {
		relative, err := filepath.Rel(request.Root, source.Path)
		if err == nil {
			repositoryFiles[strings.ToLower(filepath.ToSlash(relative))] = struct{}{}
		}
	}
	for _, required := range s.limits.RequiredFiles {
		required = filepath.ToSlash(filepath.Clean(required))
		if _, ok := repositoryFiles[strings.ToLower(required)]; !ok {
			findings = append(findings, fileFinding("required-file-missing", finding.Governance, "repository_policy", finding.High,
				fmt.Sprintf("Required repository file %s is missing", required), "Add the required file or update governance policy through review", required, 1))
		}
	}
	ownershipFindings, ownershipErrors := s.scanOwnershipPolicies(ctx, request)
	findings = append(findings, ownershipFindings...)
	failures = append(failures, ownershipErrors...)
	for _, source := range request.Files {
		if err := ctx.Err(); err != nil {
			return findings, append(failures, err)
		}
		relative, err := filepath.Rel(request.Root, source.Path)
		if err != nil {
			relative = source.Path
		}
		relative = filepath.ToSlash(relative)
		base := strings.ToLower(filepath.Base(relative))
		extension := strings.ToLower(filepath.Ext(base))
		if temporaryArtifact(base, extension) {
			findings = append(findings, fileFinding("temporary-artifact", finding.Quality, "repository_hygiene", finding.Medium,
				"Temporary atau build artifact ikut dalam perubahan", "Hapus artefak dan tambahkan pola yang sesuai ke .gitignore", relative, 1))
		}
		if base == "package.json" && !hasJavaScriptLockfile(knownFiles, filepath.ToSlash(filepath.Dir(relative))) {
			findings = append(findings, fileFinding("manifest-without-lockfile", finding.SupplyChain, "dependency_lock", finding.High,
				"JavaScript dependency manifest lacks lockfile in the same directory", "Commit package manager lockfile for reproducible dependency resolution", relative, 1))
		}
		if base != "dockerfile" && !strings.HasPrefix(base, "dockerfile.") && request.Mode == "full" {
			if base != "go.mod" && base != "package.json" && base != "package-lock.json" && !isGitHubWorkflow(relative) {
				continue
			}
		}
		reader, openErr := source.Open(ctx)
		if openErr != nil {
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, openErr))
			continue
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, 64*1024+1))
		if readErr != nil {
			_ = reader.Close()
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, readErr))
			continue
		}
		lineScanner := bufio.NewScanner(bytes.NewReader(content))
		line := 0
		for lineScanner.Scan() {
			line++
			text := lineScanner.Text()
			if (base == "dockerfile" || strings.HasPrefix(base, "dockerfile.")) && strings.EqualFold(strings.TrimSpace(text), "USER root") {
				findings = append(findings, fileFinding("docker-root-user", finding.Hardening, "container_security", finding.High,
					"Docker container is configured to run as root", "Use a non-root USER with minimum required permissions", relative, line))
			}
			if (base == "dockerfile" || strings.HasPrefix(base, "dockerfile.")) && dockerLatest(text) {
				findings = append(findings, fileFinding("docker-latest-tag", finding.SupplyChain, "unpinned_dependency", finding.High,
					"Docker base image uses mutable tag latest", "Pin base image to sha256 digest or immutable version tag", relative, line))
			}
			if base == "go.mod" && localGoReplace(text) {
				findings = append(findings, fileFinding("go-local-replace", finding.SupplyChain, "unpinned_dependency", finding.High,
					"go.mod uses replace directive to a local path", "Use reproducible module versions or uncommitted workspace files", relative, line))
			}
			if isGitHubWorkflow(relative) && mutableActionReference(text) {
				findings = append(findings, fileFinding("github-action-mutable-ref", finding.SupplyChain, "ci_dependency", finding.High,
					"GitHub Action uses a mutable ref", "Pin action to full commit SHA and record version tag in comment", relative, line))
			}
			if request.Mode != "full" && strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT") {
				findings = append(findings, fileFinding("generated-file-changed", finding.Quality, "generated_code", finding.Low,
					"Generated file is included in changes", "Ensure file is regenerated from source generator, not edited manually", relative, line))
				break
			}
		}
		if base == "package.json" {
			findings = append(findings, unpinnedPackageDependencies(content, relative)...)
			findings = append(findings, dependencyPolicyFindings(content, relative, s.limits)...)
		}
		if base == "package-lock.json" {
			findings = append(findings, licensePolicyFindings(content, relative, s.limits)...)
		}
		if scanErr := lineScanner.Err(); scanErr != nil {
			failures = append(failures, fmt.Errorf("inspect %s: %w", relative, scanErr))
		}
		if closeErr := reader.Close(); closeErr != nil {
			failures = append(failures, fmt.Errorf("close %s: %w", relative, closeErr))
		}
	}
	return findings, failures
}

func (s *Scanner) scanOwnershipPolicies(ctx context.Context, request scanner.Request) ([]finding.Finding, []error) {
	if len(s.limits.OwnershipRules) == 0 {
		return nil, nil
	}
	ownershipFile := s.limits.OwnershipFile
	if ownershipFile == "" {
		ownershipFile = "CODEOWNERS"
	}
	ownershipFile = filepath.ToSlash(filepath.Clean(ownershipFile))
	var selected *scanner.Source
	for index := range request.RepositoryFiles {
		relative, err := filepath.Rel(request.Root, request.RepositoryFiles[index].Path)
		if err == nil && strings.EqualFold(filepath.ToSlash(relative), ownershipFile) {
			selected = &request.RepositoryFiles[index]
			break
		}
	}
	if selected == nil {
		return []finding.Finding{fileFinding("ownership-file-missing", finding.Governance, "ownership", finding.High,
			fmt.Sprintf("Ownership policy file %s is missing", ownershipFile), "Add the ownership file with the configured sensitive path rules", ownershipFile, 1)}, nil
	}
	reader, err := selected.Open(ctx)
	if err != nil {
		return nil, []error{fmt.Errorf("inspect ownership file %s: %w", ownershipFile, err)}
	}
	defer reader.Close()
	type declaration struct {
		owners map[string]struct{}
		line   int
	}
	declarations := make(map[string]declaration)
	lineScanner := bufio.NewScanner(io.LimitReader(reader, 256*1024))
	line := 0
	for lineScanner.Scan() {
		line++
		text := strings.TrimSpace(lineScanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) < 2 {
			continue
		}
		owners := make(map[string]struct{}, len(fields)-1)
		for _, owner := range fields[1:] {
			owners[owner] = struct{}{}
		}
		declarations[fields[0]] = declaration{owners: owners, line: line}
	}
	if err := lineScanner.Err(); err != nil {
		return nil, []error{fmt.Errorf("inspect ownership file %s: %w", ownershipFile, err)}
	}
	var findings []finding.Finding
	for _, policy := range s.limits.OwnershipRules {
		declared, exists := declarations[policy.Path]
		var missing []string
		for _, owner := range policy.Owners {
			if _, found := declared.owners[owner]; !found {
				missing = append(missing, owner)
			}
		}
		if exists && len(missing) == 0 {
			continue
		}
		severity := policy.Severity
		if severity == "" {
			severity = finding.High
		}
		locationLine := declared.line
		if locationLine == 0 {
			locationLine = 1
		}
		item := fileFinding("sensitive-path-ownership", finding.Governance, "ownership", severity,
			fmt.Sprintf("Sensitive path %s is not assigned to all required owners", policy.Path),
			"Add the exact path pattern and required owners to the configured ownership file", ownershipFile, locationLine)
		item.Metadata = map[string]string{"sensitive_path": policy.Path, "required_owners": strings.Join(policy.Owners, ","), "missing_owners": strings.Join(missing, ",")}
		findings = append(findings, item)
	}
	return findings, nil
}

func dependencyPolicyFindings(content []byte, path string, limits Limits) []finding.Finding {
	var manifest struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}
	if json.Unmarshal(content, &manifest) != nil {
		return nil
	}
	names := make(map[string]struct{})
	for _, group := range []map[string]string{manifest.Dependencies, manifest.DevDependencies, manifest.OptionalDependencies} {
		for name := range group {
			names[name] = struct{}{}
		}
	}
	var result []finding.Finding
	for name := range names {
		denied := matchesPolicy(name, limits.DependencyDenylist)
		notAllowed := len(limits.DependencyAllowlist) > 0 && !matchesPolicy(name, limits.DependencyAllowlist)
		if !denied && !notAllowed {
			continue
		}
		item := fileFinding("dependency-policy", finding.SupplyChain, "dependency_policy", finding.High,
			fmt.Sprintf("Dependency %s is not allowed by repository policy", name), "Use an approved dependency or update policy via review", path, 1)
		item.Metadata = map[string]string{"dependency": name}
		result = append(result, item)
	}
	return result
}

func licensePolicyFindings(content []byte, path string, limits Limits) []finding.Finding {
	var lockfile struct {
		Packages map[string]struct {
			Name    string `json:"name"`
			License string `json:"license"`
		} `json:"packages"`
	}
	if json.Unmarshal(content, &lockfile) != nil {
		return nil
	}
	var result []finding.Finding
	for packagePath, pkg := range lockfile.Packages {
		if packagePath == "" || pkg.License == "" {
			continue
		}
		name := pkg.Name
		if name == "" {
			name = strings.TrimPrefix(filepath.ToSlash(packagePath), "node_modules/")
		}
		denied := matchesPolicy(pkg.License, limits.LicenseDenylist)
		notAllowed := len(limits.LicenseAllowlist) > 0 && !matchesPolicy(pkg.License, limits.LicenseAllowlist)
		if !denied && !notAllowed {
			continue
		}
		item := fileFinding("dependency-license-policy", finding.SupplyChain, "license_policy", finding.High,
			fmt.Sprintf("Dependency %s uses unapproved license %s", name, pkg.License), "Replace dependency or update license policy via legal review", path, 1)
		item.Metadata = map[string]string{"dependency": name, "license": pkg.License}
		result = append(result, item)
	}
	return result
}

func matchesPolicy(value string, patterns []string) bool {
	value = strings.ToLower(value)
	for _, pattern := range patterns {
		matched, _ := filepath.Match(strings.ToLower(pattern), value)
		if matched {
			return true
		}
	}
	return false
}

func hasJavaScriptLockfile(files map[string]struct{}, directory string) bool {
	lockfileNames := []string{"package-lock.json", "npm-shrinkwrap.json", "pnpm-lock.yaml", "yarn.lock", "bun.lock", "bun.lockb"}

	curr := directory
	for {
		prefix := curr
		if prefix == "." {
			prefix = ""
		}
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		for _, name := range lockfileNames {
			if _, ok := files[strings.ToLower(prefix+name)]; ok {
				return true
			}
		}
		if curr == "" || curr == "." {
			break
		}
		parent := filepath.ToSlash(filepath.Dir(curr))
		if parent == curr {
			break
		}
		curr = parent
	}
	return false
}

func dockerLatest(line string) bool {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < 2 || !strings.EqualFold(fields[0], "FROM") {
		return false
	}
	image := strings.ToLower(fields[1])
	return strings.HasSuffix(image, ":latest")
}

func localGoReplace(line string) bool {
	parts := strings.SplitN(strings.TrimSpace(line), "=>", 2)
	if len(parts) != 2 {
		return false
	}
	target := strings.Fields(strings.TrimSpace(parts[1]))
	return len(target) > 0 && (strings.HasPrefix(target[0], "./") || strings.HasPrefix(target[0], "../") || filepath.IsAbs(target[0]))
}

func isGitHubWorkflow(path string) bool {
	path = strings.ToLower(filepath.ToSlash(path))
	return strings.HasPrefix(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"))
}

func mutableActionReference(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "uses:") && !strings.HasPrefix(line, "- uses:") {
		return false
	}
	at := strings.LastIndex(line, "@")
	if at < 0 {
		return true
	}
	if strings.TrimSpace(line[at+1:]) == "" {
		return true
	}
	ref := strings.Fields(line[at+1:])[0]
	if len(ref) != 40 {
		return true
	}
	for _, character := range ref {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return true
		}
	}
	return false
}

func unpinnedPackageDependencies(content []byte, path string) []finding.Finding {
	var manifest struct {
		Dependencies         map[string]string      `json:"dependencies"`
		DevDependencies      map[string]string      `json:"devDependencies"`
		OptionalDependencies map[string]string      `json:"optionalDependencies"`
		PeerDependencies     map[string]string      `json:"peerDependencies"`
		Scripts              map[string]string      `json:"scripts"`
		Workspaces           interface{}            `json:"workspaces"`
	}
	if json.Unmarshal(content, &manifest) != nil {
		return nil
	}

	// collect local workspace package names to avoid false positives
	workspaceNames := map[string]bool{}
	switch ws := manifest.Workspaces.(type) {
	case []interface{}:
		for _, v := range ws {
			if s, ok := v.(string); ok {
				workspaceNames[s] = true
			}
		}
	}

	var findings []finding.Finding
	groups := []map[string]string{manifest.Dependencies, manifest.DevDependencies, manifest.OptionalDependencies, manifest.PeerDependencies}
	for _, group := range groups {
		for name, version := range group {
			value := strings.TrimSpace(version)
			lower := strings.ToLower(value)

			// skip local workspace protocol refs (workspace:, file:, link:)
			if strings.HasPrefix(lower, "workspace:") || strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, "link:") {
				continue
			}
			// skip local workspace package names
			if workspaceNames[name] {
				continue
			}

			var reason string
			switch {
			case lower == "*" || lower == "latest" || lower == "x":
				reason = "wildcard or latest version"
			case strings.HasPrefix(lower, "git+") || strings.HasPrefix(lower, "github:") ||
				strings.HasPrefix(lower, "bitbucket:") || strings.HasPrefix(lower, "gitlab:"):
				// mutable git ref: check for missing commit SHA
				if !gitRefHasSHA(value) {
					reason = "mutable Git reference without commit SHA"
				}
			case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
				if !strings.Contains(lower, "#") {
					reason = "unpinned URL dependency without hash"
				}
			}
			if reason == "" {
				continue
			}
			item := fileFinding("javascript-unpinned-dependency", finding.SupplyChain, "unpinned_dependency", finding.High,
				fmt.Sprintf("Dependency %s uses unpinned version reference (%s)", name, reason),
				"Pin dependency to a controlled version range and commit lockfile", path, 1)
			item.Metadata = map[string]string{"dependency": name, "version": version, "reason": reason}
			findings = append(findings, item)
		}
	}

	// Suspicious lifecycle scripts
	suspiciousScriptPatterns := []string{"curl", "wget", "bash", "sh ", "eval", "exec", "node -e", "python -c"}
	for scriptName, scriptBody := range manifest.Scripts {
		bodyLower := strings.ToLower(scriptBody)
		for _, pattern := range suspiciousScriptPatterns {
			if strings.Contains(bodyLower, pattern) {
				item := fileFinding("javascript-suspicious-lifecycle-script", finding.SupplyChain, "supply_chain_risk", finding.High,
					fmt.Sprintf("Lifecycle script '%s' contains suspicious command pattern: %s", scriptName, pattern),
					"Review lifecycle scripts for supply chain risk; avoid shell execution of remote content", path, 1)
				item.Metadata = map[string]string{"script": scriptName, "pattern": pattern}
				findings = append(findings, item)
				break
			}
		}
	}

	return findings
}

// gitRefHasSHA checks whether a git dependency reference contains a full commit SHA.
func gitRefHasSHA(ref string) bool {
	// look for #<sha> or @<sha> where sha is 40 hex chars
	for _, sep := range []string{"#", "@"} {
		idx := strings.LastIndex(ref, sep)
		if idx < 0 {
			continue
		}
		candidate := strings.TrimSpace(ref[idx+1:])
		if len(candidate) == 40 {
			allHex := true
			for _, c := range candidate {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					allHex = false
					break
				}
			}
			if allHex {
				return true
			}
		}
	}
	return false
}


func temporaryArtifact(base, extension string) bool {
	switch extension {
	case ".bak", ".class", ".dump", ".exe", ".orig", ".out", ".swp", ".tmp":
		return true
	}
	return strings.HasSuffix(base, "~") || base == "core"
}

func fileFinding(id string, domain finding.Domain, category string, severity finding.Severity, description, recommendation, path string, line int) finding.Finding {
	return finding.Finding{RuleID: id, Tool: "pattern", Domain: domain, Category: category, Severity: severity,
		Description: description, Recommendation: recommendation, Location: finding.Location{File: path, Line: line}}
}

type outcome struct {
	findings []finding.Finding
	err      error
}

func feedSources(ctx context.Context, sources []scanner.Source, jobs chan<- scanner.Source) {
	defer close(jobs)
	for _, source := range sources {
		select {
		case jobs <- source:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scanner) scanSource(ctx context.Context, source scanner.Source, root string) ([]finding.Finding, error) {
	file, err := source.Open(ctx)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	relative, err := filepath.Rel(root, source.Path)
	if err != nil {
		relative = source.Path
	}
	relative = filepath.ToSlash(relative)
	headers := s.headersFor(relative)
	headerMatches := make([]bool, len(headers))
	extension := strings.ToLower(filepath.Ext(source.Path))
	applicableRules := s.rulesFor(extension)
	var findings []finding.Finding
	lineNumber := 0
	counter := &countingReader{reader: io.LimitReader(file, s.limits.MaxFileBytes+1)}
	lineScanner := bufio.NewScanner(counter)
	lineScanner.Buffer(make([]byte, min(s.limits.MaxLineBytes, 64*1024)), s.limits.MaxLineBytes)
	for lineScanner.Scan() {
		lineNumber++
		line := lineScanner.Text()
		for index, header := range headers {
			maxLines := header.MaxLines
			if maxLines == 0 {
				maxLines = 20
			}
			if lineNumber <= maxLines && header.expression.MatchString(line) {
				headerMatches[index] = true
			}
		}
		if s.limits.QualityMaxLineLength > 0 && len([]rune(line)) > s.limits.QualityMaxLineLength {
			findings = append(findings, fileFinding("line-length", finding.Quality, "maintainability", finding.Low,
				fmt.Sprintf("Baris melebihi batas %d karakter", s.limits.QualityMaxLineLength),
				"Pecah expression atau data panjang menjadi struktur yang lebih mudah ditinjau", relative, lineNumber))
		}
		if counter.bytes > s.limits.MaxFileBytes {
			return findings, fmt.Errorf("%s exceeds pattern file limit of %d bytes; increase pattern_max_file_bytes", relative, s.limits.MaxFileBytes)
		}
		for _, rule := range applicableRules {
			if !rule.Regex.MatchString(line) {
				continue
			}
			findings = append(findings, finding.Finding{
				RuleID: rule.ID, Tool: s.ID(), Domain: rule.Domain, Category: rule.Category,
				Severity: rule.Severity, Description: rule.Description,
				Recommendation: rule.Recommendation, Documentation: rule.Documentation,
				Tags: append([]string(nil), rule.Tags...), Fixable: rule.Fixable,
				Snippet:  redact(rule, line),
				Location: finding.Location{File: relative, Line: lineNumber},
			})
		}
	}
	if s.limits.QualityMaxFileBytes > 0 && counter.bytes > s.limits.QualityMaxFileBytes {
		findings = append(findings, fileFinding("source-file-size", finding.Quality, "maintainability", finding.Medium,
			fmt.Sprintf("Source file melebihi batas %d bytes", s.limits.QualityMaxFileBytes),
			"Pisahkan file berdasarkan tanggung jawab atau pindahkan data besar ke format yang sesuai", relative, 1))
	}
	for index, header := range headers {
		if headerMatches[index] {
			continue
		}
		severity := header.Severity
		if severity == "" {
			severity = finding.Medium
		}
		description := header.Description
		if description == "" {
			description = fmt.Sprintf("Required header %s is missing", header.ID)
		}
		findings = append(findings, fileFinding(header.ID, finding.Governance, "required_header", severity,
			description, header.Recommendation, relative, 1))
	}
	if err := lineScanner.Err(); err != nil {
		if strings.Contains(err.Error(), "token too long") {
			return findings, fmt.Errorf("%s contains a line exceeding %d bytes; increase pattern_max_line_bytes", relative, s.limits.MaxLineBytes)
		}
		return findings, err
	}
	return findings, nil
}

func (s *Scanner) headersFor(path string) []compiledHeader {
	var result []compiledHeader
	for _, header := range s.headers {
		for _, pattern := range header.Paths {
			if matched, _ := pathpkg.Match(pattern, path); matched {
				result = append(result, header)
				break
			}
		}
	}
	return result
}

type countingReader struct {
	reader io.Reader
	bytes  int64
}

func (r *countingReader) Read(buffer []byte) (int, error) {
	read, err := r.reader.Read(buffer)
	r.bytes += int64(read)
	return read, err
}

func (s *Scanner) rulesFor(extension string) []rules.Compiled {
	result := make([]rules.Compiled, 0, len(s.genericRules)+len(s.rulesBySuffix[extension]))
	result = append(result, s.genericRules...)
	result = append(result, s.rulesBySuffix[extension]...)
	return result
}

func redact(rule rules.Compiled, value string) string {
	if rule.Category == "secret_leak" {
		return "[REDACTED: " + rule.ID + "]"
	}
	if sensitiveRule(rule) {
		return "[REDACTED: potentially sensitive source line]"
	}
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > 200 {
		value = string(runes[:200]) + "…"
	}
	return value
}

func sensitiveRule(rule rules.Compiled) bool {
	switch strings.ToLower(rule.Category) {
	case "authorization", "credential", "credentials", "personal_data", "secret", "secret_leak":
		return true
	}
	for _, tag := range rule.Tags {
		switch strings.ToLower(tag) {
		case "credential", "pii", "secret", "sensitive":
			return true
		}
	}
	return false
}

func appendMessage(current, addition string) string {
	if current == "" {
		return addition
	}
	return current + "; " + addition
}
