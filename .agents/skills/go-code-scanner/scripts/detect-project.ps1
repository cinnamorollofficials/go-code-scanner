# detect-project.ps1
# Inspects target workspace for languages, frameworks, package managers, and security-review config

param (
    [string]$TargetDir = "."
)

$ErrorActionPreference = "Stop"

$resolvedPath = Resolve-Path $TargetDir
$detections = @()
$evidence = @()

# Check Go
if (Test-Path (Join-Path $resolvedPath "go.mod")) {
    $detections += "go"
    $evidence += "Found go.mod"
}

# Check Node / JS / TS
if (Test-Path (Join-Path $resolvedPath "package.json")) {
    $detections += "javascript/typescript"
    $evidence += "Found package.json"
}

# Check Python
if ((Test-Path (Join-Path $resolvedPath "requirements.txt")) -or (Test-Path (Join-Path $resolvedPath "pyproject.toml"))) {
    $detections += "python"
    $evidence += "Found Python manifest"
}

# Check CI
$ciFound = $false
if (Test-Path (Join-Path $resolvedPath ".github/workflows")) {
    $ciFound = $true
    $evidence += "Found GitHub Actions workflows under .github/workflows"
}

# Check existing scanner config
$configFound = $false
if (Test-Path (Join-Path $resolvedPath "security-review.json")) {
    $configFound = $true
    $evidence += "Found security-review.json"
}

$output = @{
    status = "success"
    target_directory = $resolvedPath.Path
    detected_languages = $detections
    has_ci_integration = $ciFound
    has_scanner_config = $configFound
    evidence = $evidence
}

$output | ConvertTo-Json -Depth 4
