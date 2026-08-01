package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/baseline"
	cachepkg "github.com/cinnamorollofficials/go-code-scanner/cache"
	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/fixer"
	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
	"github.com/cinnamorollofficials/go-code-scanner/hook"
	"github.com/cinnamorollofficials/go-code-scanner/policy"
	"github.com/cinnamorollofficials/go-code-scanner/reporter"
	"github.com/cinnamorollofficials/go-code-scanner/rules"
	"github.com/cinnamorollofficials/go-code-scanner/suppression"
)

const version = "0.1.0-dev"

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
	case "version", "--version", "-version":
		fmt.Fprintln(stdout, version)
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
	reviewer, err := securityreview.New(cfg, securityreview.WithToolVersion(version))
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
	outputPath := cfg.Output
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(cfg.Root, outputPath)
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

func writeUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: security-review <scan|config|hook|baseline|suppress|cache|version> [options]")
}
