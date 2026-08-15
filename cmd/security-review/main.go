package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/baseline"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/buildinfo"
	cachepkg "github.com/cinnamorollofficials/go-code-scanner/pkg/cache"
	compatibilitypkg "github.com/cinnamorollofficials/go-code-scanner/pkg/compatibility"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/finding"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/fixer"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/gitrepo"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/hook"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/policy"
	releasepkg "github.com/cinnamorollofficials/go-code-scanner/pkg/release"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/reporter"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/rules"
	frontendscanner "github.com/cinnamorollofficials/go-code-scanner/pkg/scanner/frontend"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/suppression"
	"github.com/cinnamorollofficials/go-code-scanner/pkg/viewer"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return runScan(ctx, nil, stdout, stderr)
	}
	switch args[0] {
	case "scan":
		return runScan(ctx, args[1:], stdout, stderr)
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "hook":
		return runHook(ctx, args[1:], stdout, stderr)
	case "baseline":
		return runBaseline(args[1:], stdout, stderr)
	case "suppress":
		return runSuppress(args[1:], stdout, stderr)
	case "cache":
		return runCache(args[1:], stdout, stderr)
	case "ui", "view":
		return runUI(ctx, args[1:], stdout, stderr)
	case "release":
		return runRelease(args[1:], stdout, stderr)
	case "upgrade":
		return runUpgrade(args[1:], stdout, stderr)
	case "version", "--version", "-version":
		fmt.Fprintln(stdout, buildinfo.String())
		return 0
	case "help", "--help", "-h":
		writeUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		writeUsage(stderr)
		return 2
	}
}

func runScan(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to JSON configuration")
	root := flags.String("root", ".", "project root")
	output := flags.String("output", "", "JSON report path")
	changed := flags.Bool("changed", false, "scan files changed from HEAD")
	staged := flags.Bool("staged", false, "scan files staged in Git")
	ci := flags.Bool("ci", false, "fail when findings meet the threshold")
	failOn := flags.String("fail-on", "", "critical, high, medium, or low")
	quiet := flags.Bool("quiet", false, "suppress terminal summary")
	verbose := flags.Bool("verbose", false, "show scanner metadata and timing")
	color := flags.String("color", "auto", "terminal color: auto, always, or never")
	explain := flags.String("explain", "", "explain a configured rule ID and exit")
	profile := flags.String("profile", "", "scanner profile to run")
	baselinePath := flags.String("baseline", "", "finding baseline path")
	newOnly := flags.Bool("new-only", false, "apply CI policy only to findings absent from the baseline")
	format := flags.String("format", "json", "report format: json, sarif, or junit")
	fix := flags.Bool("fix", false, "apply deterministic fixes and rescan")
	dryRun := flags.Bool("dry-run", false, "preview --fix changes without writing")
	scope := flags.String("scope", "", "client scan scope: client, server, or all (default all)")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	cfg, err := loadConfig(*configPath, *root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *changed && *staged {
		fmt.Fprintln(stderr, "--changed and --staged are mutually exclusive")
		return 2
	}
	if *dryRun && !*fix {
		fmt.Fprintln(stderr, "--dry-run requires --fix")
		return 2
	}
	if *changed {
		cfg.Mode = config.ModeChanged
	}
	if *staged {
		cfg.Mode = config.ModeStaged
	}
	if *fix && cfg.Mode != config.ModeFull {
		fmt.Fprintln(stderr, "--fix is only supported for full working-tree scans")
		return 2
	}
	if *output != "" {
		cfg.Output = *output
	}
	if *profile != "" {
		cfg.SelectedProfile = *profile
	}
	if *baselinePath != "" {
		cfg.BaselineFile = *baselinePath
	}
	if *failOn != "" {
		cfg.FailOn, err = finding.ParseSeverity(*failOn)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		// An explicit CLI threshold is a global override for this invocation.
		cfg.Policy = nil
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *scope != "" {
		parsedScope, scopeErr := config.ParseScanScope(*scope)
		if scopeErr != nil {
			fmt.Fprintln(stderr, scopeErr)
			return 2
		}
		cfg.ScanScope = parsedScope
	}
	if *explain != "" {
		return explainRule(cfg, *explain, stdout, stderr)
	}
	if *format != "json" && *format != "sarif" && *format != "junit" {
		fmt.Fprintf(stderr, "invalid report format %q\n", *format)
		return 2
	}
	colorEnabled, err := terminalColor(*color, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	reviewer, err := securityreview.New(cfg, securityreview.WithToolVersion(buildinfo.String()))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := reviewer.Run(ctx)
	operationalErr := err
	if report == nil {
		fmt.Fprintln(stderr, operationalErr)
		return 3
	}
	if *fix {
		if operationalErr != nil {
			fmt.Fprintln(stderr, "refusing to apply fixes after an operational scanner failure")
			fmt.Fprintln(stderr, operationalErr)
			return 3
		}
		changes, fixErr := fixer.Apply(cfg.Root, report.Findings, *dryRun)
		if fixErr != nil {
			fmt.Fprintln(stderr, fixErr)
			return 3
		}
		for _, change := range changes {
			prefix := "Fixed"
			if *dryRun {
				prefix = "Would fix"
			}
			fmt.Fprintf(stdout, "%s %s:%d (%s)\n", prefix, change.File, change.Line, change.RuleID)
		}
		if !*dryRun && len(changes) > 0 {
			report, operationalErr = reviewer.Run(ctx)
			if report == nil {
				fmt.Fprintln(stderr, operationalErr)
				return 3
			}
		}
	}
	if *newOnly || *baselinePath != "" {
		if _, err := compareBaseline(report, cfg.Root, cfg.BaselineFile); err != nil {
			operationalErr = errors.Join(operationalErr, err)
		}
	}
	outputPath, err := config.ResolveProjectPath(cfg.Root, cfg.Output)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := writeReport(*format, outputPath, report); err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	if !*quiet {
		if err := reporter.WriteTerminalWithOptions(stdout, report, reporter.TerminalOptions{
			MaxFindings: reporter.DefaultTerminalFindingLimit, Verbose: *verbose, Color: colorEnabled,
		}); err != nil {
			fmt.Fprintln(stderr, err)
			return 3
		}
		fmt.Fprintf(stdout, "Report: %s\n", outputPath)
	}
	if operationalErr != nil {
		fmt.Fprintln(stderr, operationalErr)
		return 3
	}
	decision := policy.Evaluate(report, cfg.FailOn, cfg.Policy, *newOnly)
	if *ci && !decision.Allowed {
		return 1
	}
	return 0
}

func terminalColor(mode string, writer io.Writer) (bool, error) {
	switch mode {
	case "always":
		return true, nil
	case "never":
		return false, nil
	case "auto":
		if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
			return false, nil
		}
		file, ok := writer.(*os.File)
		if !ok {
			return false, nil
		}
		info, err := file.Stat()
		if err != nil {
			return false, nil
		}
		return info.Mode()&os.ModeCharDevice != 0, nil
	default:
		return false, fmt.Errorf("invalid color mode %q", mode)
	}
}

func explainRule(cfg config.Config, id string, stdout, stderr io.Writer) int {
	paths := make([]string, len(cfg.RuleFiles))
	for index, path := range cfg.RuleFiles {
		if !filepath.IsAbs(path) {
			path = filepath.Join(cfg.Root, path)
		}
		paths[index] = path
	}
	compiled, err := rules.Load(paths)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	for _, item := range compiled {
		if item.ID != id {
			continue
		}
		fmt.Fprintf(stdout, "Rule: %s\nDomain: %s\nCategory: %s\nSeverity: %s\nDescription: %s\n", item.ID, item.Domain, item.Category, item.Severity, item.Description)
		if item.Recommendation != "" {
			fmt.Fprintf(stdout, "Recommendation: %s\n", item.Recommendation)
		}
		if item.Documentation != "" {
			fmt.Fprintf(stdout, "Documentation: %s\n", item.Documentation)
		}
		if len(item.Tags) > 0 {
			fmt.Fprintf(stdout, "Tags: %s\n", strings.Join(item.Tags, ", "))
		}
		fmt.Fprintf(stdout, "Fixable: %t\n", item.Fixable)
		return 0
	}
	if fRule, ok := frontendscanner.LookupRule(id); ok {
		fmt.Fprintf(stdout, "Rule: %s\nDomain: %s\nCategory: %s\nSeverity: %s\nDescription: %s\n", fRule.ID, fRule.Domain, fRule.Category, fRule.Severity, fRule.Description)
		if fRule.Recommendation != "" {
			fmt.Fprintf(stdout, "Recommendation: %s\n", fRule.Recommendation)
		}
		if fRule.Documentation != "" {
			fmt.Fprintf(stdout, "Documentation: %s\n", fRule.Documentation)
		}
		if len(fRule.Tags) > 0 {
			fmt.Fprintf(stdout, "Tags: %s\n", strings.Join(fRule.Tags, ", "))
		}
		fmt.Fprintf(stdout, "Fixable: %t\n", false)
		return 0
	}
	fmt.Fprintf(stderr, "unknown rule %q\n", id)
	return 2
}

func writeReport(format, path string, report *finding.Report) error {
	switch format {
	case "json":
		return reporter.WriteJSON(path, report)
	case "sarif":
		return reporter.WriteSARIF(path, report)
	case "junit":
		return reporter.WriteJUnit(path, report)
	default:
		return fmt.Errorf("unsupported report format %q", format)
	}
}

func compareBaseline(report *finding.Report, root, path string) (baseline.Comparison, error) {
	if path == "" {
		return baseline.Comparison{}, fmt.Errorf("baseline path is required")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	file, err := baseline.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		file = &baseline.File{Version: baseline.Version, FingerprintVersion: report.FingerprintVersion}
	} else if err != nil {
		return baseline.Comparison{}, err
	}
	comparison, err := baseline.Compare(report, file)
	if err != nil {
		return baseline.Comparison{}, fmt.Errorf("compare baseline: %w", err)
	}
	return comparison, nil
}

func runBaseline(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "create" && args[0] != "update" && args[0] != "status") {
		fmt.Fprintln(stderr, "usage: security-review baseline <create|update|status> --report <path> [--baseline <path>]")
		return 2
	}
	command := args[0]
	flags := flag.NewFlagSet("baseline "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	reportPath := flags.String("report", "", "JSON report path")
	baselinePath := flags.String("baseline", ".security-baseline.json", "baseline path")
	dryRun := flags.Bool("dry-run", false, "preview baseline changes without writing")
	acceptResolved := flags.Bool("accept-resolved", false, "allow update to remove resolved baseline entries")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *reportPath == "" {
		fmt.Fprintln(stderr, "--report is required")
		return 2
	}
	report, err := loadReport(*reportPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if command == "status" {
		file, err := baseline.Load(*baselinePath)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		comparison, err := baseline.Compare(report, file)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprintf(stdout, "Baseline: new=%d existing=%d resolved=%d\n", len(comparison.New), len(comparison.Existing), len(comparison.Resolved))
		return 0
	}
	file, err := baseline.FromReport(report, time.Now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if command == "update" {
		current, loadErr := baseline.Load(*baselinePath)
		if loadErr != nil {
			fmt.Fprintln(stderr, loadErr)
			return 2
		}
		comparison, compareErr := baseline.Compare(report, current)
		if compareErr != nil {
			fmt.Fprintln(stderr, compareErr)
			return 2
		}
		fmt.Fprintf(stdout, "Baseline update preview: new=%d existing=%d resolved=%d\n", len(comparison.New), len(comparison.Existing), len(comparison.Resolved))
		if len(comparison.Resolved) > 0 && !*dryRun && !*acceptResolved {
			fmt.Fprintln(stderr, "baseline update would remove resolved findings; review with --dry-run, then pass --accept-resolved")
			return 2
		}
	}
	if *dryRun {
		fmt.Fprintf(stdout, "Baseline dry-run: %d findings would be written to %s\n", len(file.Entries), *baselinePath)
		return 0
	}
	if err := baseline.Write(*baselinePath, file); err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	fmt.Fprintf(stdout, "Baseline %s: %d findings written to %s\n", command+"d", len(file.Entries), *baselinePath)
	return 0
}

func runSuppress(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "add" {
		fmt.Fprintln(stderr, "usage: security-review suppress add --file <path> --reason <text> --expires <YYYY-MM-DD> [options]")
		return 2
	}
	flags := flag.NewFlagSet("suppress add", flag.ContinueOnError)
	flags.SetOutput(stderr)
	suppressionFile := flags.String("suppression-file", ".security-ignore", "suppression JSON path")
	file := flags.String("file", "", "finding file path")
	line := flags.Int("line", -1, "finding line, or -1 for any line")
	ruleID := flags.String("rule", "", "finding rule ID")
	fingerprint := flags.String("fingerprint", "", "finding fingerprint")
	reason := flags.String("reason", "", "reviewed suppression reason")
	expires := flags.String("expires", "", "expiry date in YYYY-MM-DD")
	ticket := flags.String("ticket", "", "approval ticket")
	approvedBy := flags.String("approved-by", "", "approver identity")
	dryRun := flags.Bool("dry-run", false, "preview without writing")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(stderr, "--file is required")
		return 2
	}
	rule := suppression.Rule{RuleID: *ruleID, Fingerprint: *fingerprint, File: *file, Line: *line, Reason: *reason, Expires: *expires, Ticket: *ticket, ApprovedBy: *approvedBy}
	result, err := suppression.Add(*suppressionFile, rule, *dryRun)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	action := "Added"
	if *dryRun {
		action = "Would add"
	}
	fmt.Fprintf(stdout, "%s suppression for %s (%d total) in %s\n", action, *file, len(result.Suppressions), *suppressionFile)
	return 0
}

func runCache(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "stats" && args[0] != "clean") {
		fmt.Fprintln(stderr, "usage: security-review cache <stats|clean> [--dir <path>]")
		return 2
	}
	command := args[0]
	flags := flag.NewFlagSet("cache "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	directory := flags.String("dir", ".go-code-scanner-cache", "cache directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	store := cachepkg.Store{Directory: *directory}
	if command == "clean" {
		removed, err := store.Clean()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 3
		}
		fmt.Fprintf(stdout, "Cache cleaned: removed=%d directory=%s\n", removed, *directory)
		return 0
	}
	stats, err := store.Stats()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	fmt.Fprintf(stdout, "Cache: entries=%d bytes=%d directory=%s\n", stats.Entries, stats.Bytes, *directory)
	return 0
}

func runRelease(args []string, stdout, stderr io.Writer) int {
	if len(args) >= 1 && args[0] == "archive" {
		return runReleaseArchive(args[1:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "checksums" && args[1] == "verify" {
		return runReleaseChecksumsVerify(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "provenance" && args[1] == "generate" {
		return runReleaseProvenanceGenerate(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "provenance" && args[1] == "sign" {
		return runReleaseProvenanceSign(args[2:], stdout, stderr)
	}
	if len(args) >= 2 && args[0] == "changelog" && args[1] == "validate" {
		return runChangelogValidate(args[2:], stdout, stderr)
	}
	if len(args) == 0 || args[0] != "verify" {
		fmt.Fprintln(stderr, "usage: security-review release <archive|checksums verify|provenance generate|provenance sign|verify|changelog validate> [options]")
		return 2
	}
	flags := flag.NewFlagSet("release verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	provenance := flags.String("provenance", "", "provenance manifest path")
	signaturePath := flags.String("signature", "", "detached base64 signature path")
	publicKeyPath := flags.String("public-key", "", "PEM Ed25519 public key path")
	directory := flags.String("directory", "", "verify provenance subjects in this directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release verify does not accept positional arguments")
		return 2
	}
	if *provenance == "" || *signaturePath == "" || *publicKeyPath == "" {
		fmt.Fprintln(stderr, "--provenance, --signature, and --public-key are required")
		return 2
	}
	signature, err := os.ReadFile(*signaturePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	publicKey, err := releasepkg.LoadPublicKey(*publicKeyPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := releasepkg.VerifyFile(*provenance, string(signature), publicKey); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *directory != "" {
		if err := releasepkg.VerifyProvenance(*provenance, *directory); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, "Provenance subjects verified")
	}
	fmt.Fprintln(stdout, "Provenance signature verified")
	return 0
}

func runReleaseProvenanceGenerate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("release provenance generate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	directory := flags.String("directory", "", "directory containing release artifacts")
	outputPath := flags.String("output", "", "output provenance JSON path")
	version := flags.String("version", "", "release version")
	commit := flags.String("commit", "", "source commit")
	buildDate := flags.String("build-date", "", "build timestamp in RFC3339 format")
	builder := flags.String("builder", "", "builder identity")
	privateKeyPath := flags.String("private-key", "", "optional PEM Ed25519 private key path")
	signaturePath := flags.String("signature", "", "optional detached signature output path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release provenance generate does not accept positional arguments")
		return 2
	}
	if *directory == "" || *outputPath == "" || *version == "" || *commit == "" || *buildDate == "" || *builder == "" {
		fmt.Fprintln(stderr, "--directory, --output, --version, --commit, --build-date, and --builder are required")
		return 2
	}
	if (*privateKeyPath == "") != (*signaturePath == "") {
		fmt.Fprintln(stderr, "--private-key and --signature must be provided together")
		return 2
	}
	stamp, err := time.Parse(time.RFC3339, *buildDate)
	if err != nil {
		fmt.Fprintln(stderr, "--build-date must be an RFC3339 timestamp")
		return 2
	}
	options := releasepkg.ProvenanceOptions{Version: *version, Commit: *commit, BuildDate: stamp, Builder: *builder}
	if err := releasepkg.WriteProvenance(*directory, *outputPath, options); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := releasepkg.VerifyProvenance(*outputPath, *directory); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "Provenance generated: %s\n", *outputPath)
	if *privateKeyPath != "" {
		if code := signProvenance(*outputPath, *privateKeyPath, *signaturePath, stdout, stderr); code != 0 {
			return code
		}
	}
	return 0
}

func runReleaseProvenanceSign(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("release provenance sign", flag.ContinueOnError)
	flags.SetOutput(stderr)
	provenancePath := flags.String("provenance", "", "provenance JSON path")
	privateKeyPath := flags.String("private-key", "", "PEM Ed25519 private key path")
	outputPath := flags.String("output", "", "detached base64 signature output path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release provenance sign does not accept positional arguments")
		return 2
	}
	if *provenancePath == "" || *privateKeyPath == "" || *outputPath == "" {
		fmt.Fprintln(stderr, "--provenance, --private-key, and --output are required")
		return 2
	}
	return signProvenance(*provenancePath, *privateKeyPath, *outputPath, stdout, stderr)
}

func signProvenance(provenancePath, privateKeyPath, outputPath string, stdout, stderr io.Writer) int {
	signature, err := releasepkg.SignFile(provenancePath, privateKeyPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	temporary, err := os.CreateTemp(filepath.Dir(outputPath), ".provenance-signature-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err == nil {
		_, err = fmt.Fprintln(temporary, signature)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(temporaryPath, outputPath)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	fmt.Fprintf(stdout, "Provenance signature created: %s\n", outputPath)
	return 0
}

func runReleaseChecksumsVerify(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("release checksums verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "SHA256SUMS manifest path")
	directory := flags.String("directory", "", "directory containing release artifacts")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release checksums verify does not accept positional arguments")
		return 2
	}
	if *manifestPath == "" || *directory == "" {
		fmt.Fprintln(stderr, "--manifest and --directory are required")
		return 2
	}
	if err := releasepkg.VerifyChecksums(*manifestPath, *directory); err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, releasepkg.ErrChecksumMismatch) {
			return 1
		}
		return 2
	}
	fmt.Fprintln(stdout, "Release checksums verified")
	return 0
}

func runReleaseArchive(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("release archive", flag.ContinueOnError)
	flags.SetOutput(stderr)
	binaryPath := flags.String("binary", "", "input binary path")
	outputPath := flags.String("output", "", "output .tar.gz or .zip path")
	timestamp := flags.String("timestamp", "", "archive timestamp in RFC3339 format")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release archive does not accept positional arguments")
		return 2
	}
	if *binaryPath == "" || *outputPath == "" || *timestamp == "" {
		fmt.Fprintln(stderr, "--binary, --output, and --timestamp are required")
		return 2
	}
	stamp, err := time.Parse(time.RFC3339, *timestamp)
	if err != nil {
		fmt.Fprintln(stderr, "--timestamp must be an RFC3339 timestamp")
		return 2
	}
	if !strings.HasSuffix(*outputPath, ".tar.gz") && !strings.HasSuffix(*outputPath, ".zip") {
		fmt.Fprintln(stderr, "--output must end in .tar.gz or .zip")
		return 2
	}
	if err := releasepkg.ArchiveBinary(*binaryPath, *outputPath, stamp); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	fmt.Fprintf(stdout, "Release archive created: %s\n", *outputPath)
	return 0
}

func runChangelogValidate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("release changelog validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("file", "website/docs/changelog.md", "changelog path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "release changelog validate does not accept positional arguments")
		return 2
	}
	data, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := releasepkg.ValidateChangelog(data); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "Changelog valid: %s\n", *path)
	return 0
}

func runUpgrade(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "check" {
		fmt.Fprintln(stderr, "usage: security-review upgrade check [--contract <path>]")
		return 2
	}
	flags := flag.NewFlagSet("upgrade check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	contractPath := flags.String("contract", "", "previous compatibility contract path")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "upgrade check does not accept positional arguments")
		return 2
	}
	current := compatibilitypkg.Current()
	if *contractPath == "" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(current); err != nil {
			fmt.Fprintln(stderr, err)
			return 3
		}
		return 0
	}
	data, err := os.ReadFile(*contractPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	previous, err := compatibilitypkg.Decode(data)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	changes := compatibilitypkg.Compare(previous, current)
	if len(changes) == 0 {
		fmt.Fprintln(stdout, "Compatibility contract unchanged")
		return 0
	}
	fmt.Fprintln(stdout, "Compatibility migration required:")
	for _, change := range changes {
		fmt.Fprintf(stdout, "- %s: %s -> %s\n", change.Field, change.From, change.To)
	}
	return 1
}

func loadReport(path string) (*finding.Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read report: %w", err)
	}
	var report finding.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("decode report: %w", err)
	}
	return &report, nil
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "validate" {
		fmt.Fprintln(stderr, "usage: security-review config validate <path>")
		return 2
	}
	cfg, err := config.Load(args[1])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	fmt.Fprintf(stdout, "configuration valid for %s\n", cfg.Project)
	return 0
}

func runHook(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: security-review hook <install|uninstall|status|run> [options]")
		return 2
	}
	if args[0] == "run" {
		return runHookEvent(ctx, args[1:], stdout, stderr)
	}
	if args[0] != "install" && args[0] != "uninstall" && args[0] != "status" {
		fmt.Fprintf(stderr, "unknown hook command %q\n", args[0])
		return 2
	}
	event := hook.PreCommit
	flagArgs := args[1:]
	if len(flagArgs) > 0 && !strings.HasPrefix(flagArgs[0], "-") {
		event = flagArgs[0]
		flagArgs = flagArgs[1:]
	}
	if !hook.ValidEvent(event) {
		fmt.Fprintf(stderr, "unsupported hook %q\n", event)
		return 2
	}
	flags := flag.NewFlagSet("hook "+args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "path inside the Git repository")
	if err := flags.Parse(flagArgs); err != nil {
		return 2
	}
	repository, err := gitrepo.Open(ctx, *root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	binary, err := os.Executable()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	manager, err := hook.NewManager(repository, binary)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	switch args[0] {
	case "install":
		err = manager.Install(ctx, event)
	case "uninstall":
		err = manager.Uninstall(ctx, event)
	case "status":
		var state hook.State
		state, err = manager.Status(ctx, event)
		if err == nil {
			fmt.Fprintf(stdout, "%s: %s\n", event, state)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	return 0
}

func runHookEvent(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || !hook.ValidEvent(args[0]) {
		fmt.Fprintln(stderr, "usage: security-review hook run <pre-commit|commit-msg|pre-push> [options]")
		return 2
	}
	event := args[0]
	flags := flag.NewFlagSet("hook run "+event, flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "path inside the Git repository")
	messageFile := flags.String("file", "", "commit message file (commit-msg only)")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	repository, err := gitrepo.Open(ctx, *root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	configPath := filepath.Join(repository.Root(), "security-review.json")
	cfg, err := loadConfig("", repository.Root())
	hasConfig := false
	if info, statErr := os.Stat(configPath); statErr == nil && info.Mode().IsRegular() {
		hasConfig = true
		cfg, err = loadConfig(configPath, repository.Root())
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	configuredHook := configuredHookForEvent(cfg.Hooks, event)
	if !configuredHook.Enabled {
		fmt.Fprintf(stdout, "%s: disabled\n", event)
		return 0
	}
	if event == hook.CommitMsg {
		if *messageFile == "" {
			fmt.Fprintln(stderr, "commit-msg requires --file")
			return 2
		}
		content, readErr := os.ReadFile(*messageFile)
		if readErr != nil {
			fmt.Fprintf(stderr, "read commit message: %v\n", readErr)
			return 3
		}
		if validateErr := hook.ValidateCommitMessage(string(content), configuredHook.MessagePattern, configuredHook.MaxSubjectLength); validateErr != nil {
			fmt.Fprintf(stderr, "commit-msg: %v\n", validateErr)
			return 1
		}
		fmt.Fprintln(stdout, "commit-msg: valid")
		return 0
	}
	scanArgs := []string{"--root", repository.Root(), "--ci", "--profile", configuredHook.Profile}
	if hasConfig {
		scanArgs = []string{"--config", configPath, "--ci", "--profile", configuredHook.Profile}
	}
	if configuredHook.StagedOnly {
		scanArgs = append(scanArgs, "--staged")
	}
	if configuredHook.NewOnly {
		scanArgs = append(scanArgs, "--new-only")
	}
	return runScan(ctx, scanArgs, stdout, stderr)
}

func configuredHookForEvent(hooks config.Hooks, event string) config.Hook {
	switch event {
	case hook.CommitMsg:
		return hooks.CommitMsg
	case hook.PrePush:
		return hooks.PrePush
	default:
		return hooks.PreCommit
	}
}

func loadConfig(path, root string) (config.Config, error) {
	if path != "" {
		return config.Load(path)
	}
	cfg := config.Default()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return config.Config{}, err
	}
	cfg.Root = absRoot
	return cfg, cfg.Validate()
}

func runUI(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ui", flag.ContinueOnError)
	flags.SetOutput(stderr)
	port := flags.Int("port", 8080, "HTTP server port")
	host := flags.String("host", "127.0.0.1", "HTTP server host")
	root := flags.String("root", ".", "project root directory to scan")
	configPath := flags.String("config", "", "path to JSON configuration file")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(stderr, "invalid root path: %v\n", err)
		return 2
	}

	server := viewer.NewServer(viewer.ServerOptions{
		Root:       absRoot,
		ConfigPath: *configPath,
	})

	addr := fmt.Sprintf("%s:%d", *host, *port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Handler(),
	}

	fmt.Fprintf(stdout, "\n🛡️  Go Code Scanner Dashboard is running at http://%s\n", addr)
	fmt.Fprintf(stdout, "📁 Workspace: %s\n", absRoot)
	fmt.Fprintln(stdout, "Press Ctrl+C to stop the dashboard server.")

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		return 0
	case err := <-serverErr:
		fmt.Fprintf(stderr, "server error: %v\n", err)
		return 1
	}
}

func writeUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: security-review <scan|ui|config|hook|baseline|suppress|cache|release|upgrade|version> [options]")
}

