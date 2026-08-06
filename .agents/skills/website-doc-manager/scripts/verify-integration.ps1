# verify-integration.ps1 - Integration verification script for go-code-scanner website
$ErrorActionPreference = "Stop"

Write-Host "==> Verifying Website Integration and Documentation Build" -ForegroundColor Cyan

$WorkspaceRoot = Get-Location

# 1. Regenerate Rule Catalog if Go generator exists
$GenRuleScript = Join-Path $WorkspaceRoot "cmd\gen-rule-catalog\main.go"
if (Test-Path $GenRuleScript) {
    Write-Host "--> Running rule catalog generator..." -ForegroundColor Yellow
    go run ./cmd/gen-rule-catalog
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Rule catalog generator failed"
        exit 1
    }
    Write-Host "[OK] Rule catalog regenerated successfully" -ForegroundColor Green
}

# 2. Build VitePress documentation
Write-Host "--> Building VitePress site (npm run docs:build)..." -ForegroundColor Yellow
$WebsiteDir = Join-Path $WorkspaceRoot "website"

cmd /c "npm --prefix `"$WebsiteDir`" run docs:build"
if ($LASTEXITCODE -ne 0) {
    Write-Error "VitePress documentation build failed"
    exit 1
}

Write-Host "[OK] VitePress site built successfully with 0 broken links!" -ForegroundColor Green
Write-Host "==> Integration verification complete" -ForegroundColor Cyan
