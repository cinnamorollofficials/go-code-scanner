---
title: Scanner Definitions Configuration
description: "For configuration authors: look up pattern, command, and adapter scanner declarations and validation rules."
---

# Scanner Definitions Configuration

Declare custom command-line scanners or configure one of the supported external
tool adapters. The required built-in `pattern` scanner is registered
automatically and does not need a scanner declaration.

## Scanner Object Schema

```json
{
  "scanners": {
    "custom-linter": {
      "enabled": true,
      "required": false,
      "type": "command",
      "command": ["custom-linter", "--json"],
      "domain": "security",
      "severity": "HIGH",
      "category": "static_analysis",
      "description": "Custom linter reported a security issue",
      "workspace": "root",
      "on_missing": "skip",
      "timeout": "30s"
    }
  }
}
```

### Core Fields

| Field | Required? | Allowed values / default | Description |
| :--- | :---: | :--- | :--- |
| `enabled` | No | `false` | Makes the scanner eligible to run when its ID is selected by a profile. |
| `required` | No | `false` | Treats execution failures as operational scan failures. The CLI still writes the report, then returns exit code `3`. Optional failures are reported as warnings. |
| `type` | **Yes** | `command`, `adapter` | Selects a custom command or a supported adapter. `builtin` and `external` are not valid values. |
| `timeout` | No | Positive Go duration such as `30s` or `2m` | Limits scanner execution time. |
| `workspace` | No | `root` (default), `staged` | Runs in the project root or an isolated snapshot of staged files. |
| `on_missing` | No | `fail` (default), `skip` | Controls whether a missing executable fails or skips the scanner. |
| `environment` | No | List of environment variable names | Adds explicitly named variables to the scanner's restricted environment. |

### Command Scanner Fields

| Field | Required? | Allowed values / default | Description |
| :--- | :---: | :--- | :--- |
| `command` | **Yes** | Executable followed by arguments | Defines the command. The executable must be a PATH name or a clean absolute path. |
| `domain` | **Yes** | `quality`, `reliability`, `hardening`, `security`, `supply_chain`, `governance` | Assigns findings to a canonical policy domain. |
| `severity` | **Yes** | `CRITICAL`, `HIGH`, `MEDIUM`, `LOW` | Sets the severity for exit-code and path-based findings. |
| `category` | **Yes** | Non-empty string | Classifies the finding within its domain. |
| `description` | **Yes** | Non-empty string | Explains what the command reported. |
| `finding_exit_codes` | No | `[1]` | Exit codes that mean findings rather than operational failure. Exit code `0` cannot represent findings. |
| `output_format` | No | `exit-code` (default), `json-lines`, `paths` | Selects how scanner output is converted to findings. |

### Adapter Scanner Example

Adapters supply their own command, domain, severity, category, parser, and
description. Reference one by its supported adapter name:

```json
{
  "scanners": {
    "govulncheck": {
      "enabled": true,
      "required": false,
      "type": "adapter",
      "adapter": "govulncheck",
      "on_missing": "skip",
      "timeout": "2m"
    }
  }
}
```

Add the scanner ID to a profile before selecting that profile for a scan. See
[Profiles and Policy](/reference/config/profiles-and-policy) and the
[Scanner Compatibility Reference](/reference/scanners) for supported adapters.
