# Configuration Schema & Options

`security-review.json` controls scan execution, domain severity thresholds, external tool adapters, and policy gates.

## Example Configuration

```json
{
  "$schema": "https://raw.githubusercontent.com/cinnamorollofficials/go-code-scanner/main/docs/schema/config.schema.json",
  "root": ".",
  "profile": "fast",
  "fail_on": "high",
  "policy": {
    "security": "high",
    "hardening": "medium",
    "reliability": "high",
    "quality": "low",
    "supply_chain": "high",
    "governance": "high"
  },
  "scanners": {
    "sqltaint": {
      "enabled": true,
      "required": true
    }
  }
}
```

## Profiles

- `fast`: Offline built-in checks suitable for local pre-commit hooks.
- `standard`: Broader checks suitable for pre-push verification.
- `full`: Complete CI scanner set including external tool adapters.
- `frontend`: Browser client & JS/TS framework checks.
