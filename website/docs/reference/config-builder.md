---
title: Config Builder
description: "For configuration authors: create security-review.json locally with browser controls and CLI-validated presets."
---

# Config Builder

Use the interactive builder below to customize and export a
`security-review.json` file. Configuration state remains in the browser unless
you explicitly copy or download it.

The nine presets are executable starting points rather than decorative examples:

- **Offline** defines an `offline` profile containing only the built-in pattern
  scanner. Run it with `security-review scan --profile offline`.
- **External Scanner** configures the supported Gitleaks adapter and an
  `external` profile. Missing Gitleaks installations are skipped by default.
- **Gradual Adoption** sets the baseline path and enables new-only evaluation in
  the staged pre-commit hook.

Browser validation provides immediate feedback for common schema, enum, path,
profile, policy, frontend, cache, and scanner errors. After downloading a file,
run the authoritative product validator before committing it:

```sh
security-review config validate security-review.json
```

<ConfigBuilder />
