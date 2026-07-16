# Contract: Stream Separation (F07)

**Feature**: `007-scriptability-exits` | **Date**: 2026-07-16
**DoD commands**: `config show`, `config validate`, misuse paths via root flags

---

## Rules

1. **Success (normal, non-debug)**
   - Useful payload / success message → **stdout**
   - **stderr** MUST be empty after `strings.TrimSpace`

2. **Failure**
   - User-facing message → **stderr** (MAY be prefixed with `Ошибка:` by `cli.Execute`)
   - **stdout** MUST be empty after trim
   - Process exit via `cli.ExitCode(err)` ≠ 0

3. **Help / version success**
   - Text on **stdout**, stderr empty, exit `0`

4. **Debug/verbose**
   - Out of mandatory DoD; if added later, diagnostics MAY use stderr without moving data payload off stdout

---

## Harness

```text
stdout, stderr, err := executeForTest(args)
code := ExitCode(err)
```

Assertions for DoD:

| Scenario | err | code | stdout | stderr |
|----------|-----|------|--------|--------|
| `config show` OK | nil | 0 | non-empty data | empty |
| `config show` missing config | non-nil | ≠0 (typically 1) | empty | contains explanation |
| `config validate` OK | nil | 0 | success text | empty |
| `config validate` no token | non-nil | 2 | empty | set-token hint |
| `config validate` mock 404 | non-nil | 3 | empty | not found |
| `config validate` mock 401 | non-nil | 1 | empty | catalog |
| `--unknown-flag` | non-nil | 1 | empty | non-empty |

---

## Non-goals

- Buffering framework for streaming list commands (F08+)
- Changing F06 color policy (ANSI still forbidden on non-TTY data stdout)
