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
	fmt.Fprintln(writer, "usage: security-review <scan|config|version> [options]")
}
