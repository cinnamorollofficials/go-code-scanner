# detect-project.ps1 - Detection script for go-code-scanner website environment
$ErrorActionPreference = "Stop"

Write-Host "==> Detecting go-code-scanner Website Environment" -ForegroundColor Cyan

# 1. Check workspace root and website directory
$WorkspaceRoot = Get-Location
$WebsiteDir = Join-Path $WorkspaceRoot "website"

if (-not (Test-Path $WebsiteDir)) {
    Write-Error "Website directory not found at $WebsiteDir"
    exit 1
}

$PackageJson = Join-Path $WebsiteDir "package.json"
if (-not (Test-Path $PackageJson)) {
    Write-Error "package.json not found in website directory"
    exit 1
}

Write-Host "[OK] Website directory and package.json found" -ForegroundColor Green

# 2. Check Node.js and NPM
try {
    $nodeVersion = node --version
    Write-Host "[OK] Node.js version: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Warning "[WARN] Node.js is not installed or not in PATH"
}

# 3. Check Go runtime
try {
    $goVersion = go version
    Write-Host "[OK] Go version: $goVersion" -ForegroundColor Green
} catch {
    Write-Warning "[WARN] Go runtime is not installed or not in PATH"
}

# 4. Check Docs directory structure
$DocsDir = Join-Path $WebsiteDir "docs"
if (Test-Path $DocsDir) {
    $mdCount = (Get-ChildItem -Path $DocsDir -Recurse -Filter "*.md").Count
    Write-Host "[OK] Docs directory found with $mdCount markdown files" -ForegroundColor Green
} else {
    Write-Error "Docs directory not found at $DocsDir"
    exit 1
}

Write-Host "==> Project environment detection complete" -ForegroundColor Cyan
