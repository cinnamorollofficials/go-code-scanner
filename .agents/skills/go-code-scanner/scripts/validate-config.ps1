# validate-config.ps1
# Validates security-review configuration file using native CLI or JSON schema rules

param (
    [string]$ConfigFile = "security-review.json"
)

$ErrorActionPreference = "Continue"

if (-not (Test-Path $ConfigFile)) {
    $result = @{
        status = "error"
        message = "Configuration file '$ConfigFile' not found"
    }
    $result | ConvertTo-Json
    exit 1
}

# Try running native security-review CLI if available
$cliName = "security-review"
$cliPath = Get-Command $cliName -ErrorAction SilentlyContinue

if ($cliPath) {
    $cliOutput = & $cliName config validate $ConfigFile 2>&1
    if ($LASTEXITCODE -eq 0) {
        $result = @{
            status = "valid"
            file = $ConfigFile
            validator = "security-review-cli"
            output = "$cliOutput"
        }
        $result | ConvertTo-Json
        exit 0
    } else {
        $result = @{
            status = "invalid"
            file = $ConfigFile
            validator = "security-review-cli"
            error = "$cliOutput"
        }
        $result | ConvertTo-Json
        exit 1
    }
} else {
    # Basic JSON syntax validation fallback
    try {
        $raw = Get-Content $ConfigFile -Raw
        $json = $raw | ConvertFrom-Json
        $result = @{
            status = "valid"
            file = $ConfigFile
            validator = "powershell-json-fallback"
            message = "Valid JSON structure (CLI unavailable for full rule validation)"
        }
        $result | ConvertTo-Json
        exit 0
    } catch {
        $result = @{
            status = "invalid"
            file = $ConfigFile
            validator = "powershell-json-fallback"
            error = $_.Exception.Message
        }
        $result | ConvertTo-Json
        exit 1
    }
}
