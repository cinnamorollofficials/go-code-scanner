# verify-integration.ps1
# Runs a dry-run or verification scan using the security-review CLI

param (
    [string]$TargetDir = ".",
    [string]$Profile = "fast",
    [switch]$Staged = $false
)

$ErrorActionPreference = "Continue"

$cliName = "security-review"
$cliPath = Get-Command $cliName -ErrorAction SilentlyContinue

if (-not $cliPath) {
    # Check if local binary exists in workspace
    $localBin = Join-Path $TargetDir "security-review.exe"
    if (Test-Path $localBin) {
        $cliName = $localBin
    } else {
        $result = @{
            status = "blocked"
            message = "security-review CLI binary not found in PATH or working directory. Please build using 'go build ./cmd/security-review'."
        }
        $result | ConvertTo-Json
        exit 2
    }
}

$args = @("scan", "--root", $TargetDir, "--profile", $Profile)
if ($Staged) {
    $args += "--staged"
}

$output = & $cliName @args 2>&1
$exitCode = $LASTEXITCODE

if ($exitCode -eq 0) {
    $result = @{
        status = "passed"
        exit_code = 0
        message = "Scan completed cleanly with no policy violations."
        raw_output = "$output"
    }
    $result | ConvertTo-Json
    exit 0
} elseif ($exitCode -eq 1) {
    $result = @{
        status = "failed"
        exit_code = 1
        message = "Scan completed with policy violations or security findings."
        raw_output = "$output"
    }
    $result | ConvertTo-Json
    exit 1
} else {
    $result = @{
        status = "blocked"
        exit_code = $exitCode
        message = "Operational error during scan execution."
        raw_output = "$output"
    }
    $result | ConvertTo-Json
    exit $exitCode
}
