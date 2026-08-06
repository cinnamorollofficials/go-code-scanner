# validate-config.ps1 - Validation script for VitePress config and doc links
$ErrorActionPreference = "Stop"

Write-Host "==> Validating VitePress Configuration & Document References" -ForegroundColor Cyan

$WorkspaceRoot = Get-Location
$WebsiteDir = Join-Path $WorkspaceRoot "website"
$DocsDir = Join-Path $WebsiteDir "docs"
$VitepressConfig = Join-Path $DocsDir ".vitepress\config.mts"

if (-not (Test-Path $VitepressConfig)) {
    Write-Error "VitePress configuration file not found at $VitepressConfig"
    exit 1
}

Write-Host "[OK] VitePress config found: $VitepressConfig" -ForegroundColor Green

# Extract text links matching pattern 'link: "/..."' or 'link: "/..."'
$configContent = Get-Content $VitepressConfig -Raw
$linkRegex = "link:\s*['""](/[^'""]+)['""]"
$matches = [regex]::Matches($configContent, $linkRegex)

$missingCount = 0
foreach ($match in $matches) {
    $linkPath = $match.Groups[1].Value.TrimEnd('/')
    
    # Handle root or path mapping to .md file or index.md
    if ($linkPath -eq "" -or $linkPath -eq "/") {
        $targetFile = Join-Path $DocsDir "index.md"
    } else {
        $relativeMd = $linkPath.TrimStart('/') + ".md"
        $targetFile = Join-Path $DocsDir $relativeMd
        if (-not (Test-Path $targetFile)) {
            $relativeIndex = Join-Path ($linkPath.TrimStart('/')) "index.md"
            $targetFile = Join-Path $DocsDir $relativeIndex
        }
    }

    if (-not (Test-Path $targetFile)) {
        Write-Host "[FAIL] Config link '$linkPath' target file not found: $targetFile" -ForegroundColor Red
        $missingCount++
    }
}

if ($missingCount -eq 0) {
    Write-Host "[OK] All $($matches.Count) config sidebar/nav link targets exist!" -ForegroundColor Green
} else {
    Write-Error "Found $missingCount broken link target(s) in VitePress config"
    exit 1
}

Write-Host "==> VitePress config validation complete" -ForegroundColor Cyan
