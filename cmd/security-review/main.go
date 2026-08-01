package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	securityreview "github.com/cinnamorollofficials/go-code-scanner"
	"github.com/cinnamorollofficials/go-code-scanner/config"
	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
	"github.com/cinnamorollofficials/go-code-scanner/hook"
	"github.com/cinnamorollofficials/go-code-scanner/policy"
	"github.com/cinnamorollofficials/go-code-scanner/reporter"
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
	if *changed {
		cfg.Mode = config.ModeChanged
	}
	if *staged {
		cfg.Mode = config.ModeStaged
	}
	if *output != "" {
		cfg.Output = *output
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
	reviewer, err := securityreview.New(cfg)
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
	outputPath := cfg.Output
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(cfg.Root, outputPath)
	}
	if err := reporter.WriteJSON(outputPath, report); err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	if !*quiet {
		if err := reporter.WriteTerminal(stdout, report); err != nil {
			fmt.Fprintln(stderr, err)
			return 3
		}
		fmt.Fprintf(stdout, "Report: %s\n", outputPath)
	}
	if operationalErr != nil {
		fmt.Fprintln(stderr, operationalErr)
		return 3
	}
	if *ci && len(policy.ViolationsByDomain(report, cfg.FailOn, cfg.Policy)) > 0 {
		return 1
	}
	return 0
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
	flags := flag.NewFlagSet("hook "+args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "path inside the Git repository")
	if err := flags.Parse(args[1:]); err != nil {
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
		err = manager.Install(ctx, hook.PreCommit)
	case "uninstall":
		err = manager.Uninstall(ctx, hook.PreCommit)
	case "status":
		var state hook.State
		state, err = manager.Status(ctx, hook.PreCommit)
		if err == nil {
			fmt.Fprintf(stdout, "%s: %s\n", hook.PreCommit, state)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	return 0
}

func runHookEvent(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != hook.PreCommit {
		fmt.Fprintln(stderr, "usage: security-review hook run pre-commit [--root <path>]")
		return 2
	}
	flags := flag.NewFlagSet("hook run pre-commit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "path inside the Git repository")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	repository, err := gitrepo.Open(ctx, *root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	scanArgs := []string{"--root", repository.Root(), "--staged", "--ci"}
	configPath := filepath.Join(repository.Root(), "security-review.json")
	if info, statErr := os.Stat(configPath); statErr == nil && info.Mode().IsRegular() {
		scanArgs = []string{"--config", configPath, "--staged", "--ci"}
	}
	return runScan(ctx, scanArgs, stdout, stderr)
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
	fmt.Fprintln(writer, "usage: security-review <scan|config|hook|version> [options]")
}
