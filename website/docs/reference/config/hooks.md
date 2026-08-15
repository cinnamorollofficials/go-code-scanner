---
title: Git Hooks Configuration
description: "For configuration authors: look up pre-commit, commit-msg, and pre-push Git hook fields and defaults."
---

# Git Hooks Configuration

Configure automated git hook execution rules for pre-commit, commit-msg, and pre-push triggers.

## Schema

```json
{
  "hooks": {
    "pre_commit": {
      "enabled": true,
      "profile": "fast",
      "staged_only": true,
      "new_only": true
    },
    "commit_msg": {
      "enabled": true,
      "message_pattern": "^(feat|fix|docs|style|refactor|test|ci|chore)(\\(.+\\))?: .+",
      "max_subject_length": 72
    }
  }
}
```

### Hook Object Fields

- **`enabled`** (`bool`): Enables or disables hook execution.
- **`profile`** (`string`): Performance profile assigned to hook execution (default: `"fast"` for pre-commit).
- **`staged_only`** (`bool`): Restricts scan discovery strictly to git-staged index snapshot.
- **`new_only`** (`bool`): Evaluates findings relative to baseline snapshot.
- **`message_pattern`** (`string`): Regular expression pattern validating commit message subject line format.
- **`max_subject_length`** (`int`): Maximum character length for first line of commit message.
