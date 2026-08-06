---
title: Profiles & Policy Configuration
description: Field reference for performance profiles, domain policy thresholds, and failure modes.
---

# Profiles & Policy Configuration

Configure performance profiles, severity gates, and domain policy thresholds.

## Field Reference

### `mode` (`string`)
- **Allowed Values**: `"full"`, `"changed"`, `"staged"`
- **Default**: `"full"`
- **Description**: Default file discovery mode.

### `fail_on` (`string`)
- **Allowed Values**: `"CRITICAL"`, `"HIGH"`, `"MEDIUM"`, `"LOW"`, `"INFO"`
- **Default**: `"CRITICAL"`
- **Description**: Global minimum severity required to trigger exit code `1` under policy enforcement.

### `profiles` (`map[string]string[]`)
- **Description**: Map of profile names (`fast`, `standard`, `full`, `frontend`) to lists of enabled scanner IDs.

### `policy` (`map[Domain]Severity`)
- **Description**: Per-domain policy override mapping finding domains (`security`, `secrets`, `vulnerabilities`, `governance`, `architecture`, `frontend`) to minimum fail severities.
