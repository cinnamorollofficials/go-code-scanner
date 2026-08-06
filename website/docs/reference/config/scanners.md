---
title: Scanner Definitions Configuration
description: Field reference for built-in, external command, and adapter scanner declarations.
---

# Scanner Definitions Configuration

Declare built-in, command-line, or external adapter scanners.

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
      "timeout": "30s"
    }
  }
}
```

### Fields

- **`enabled`** (`bool`): Enables or disables execution.
- **`required`** (`bool`): If `true`, missing scanner executables trigger exit code `2`.
- **`type`** (`string`): Scanner implementation type (`builtin`, `command`, `adapter`).
- **`command`** (`string[]`): Executable binary and argument list for `command` scanners.
- **`timeout`** (`string`): Execution timeout duration (e.g. `"30s"`, `"2m"`).
- **`domain`** (`string`): Target finding domain (`security`, `secrets`, `vulnerabilities`, `governance`, `architecture`, `frontend`).
