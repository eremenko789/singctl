# Contract: CLI Exit Codes & Validate Wiring (F05)

**Feature**: `005-error-retry` | **Date**: 2026-07-16
**Extends**: `specs/004-api-adapter-auth/contracts/cli-config-validate.md`

---

## ExitCode helper

```text
cli.ExitCode(err error) int
```

| Condition | Code |
|-----------|------|
| `err == nil` | `0` |
| Kind NotFound (HTTP 404 classified) | `3` |
| Kind Config (missing token / factory config) | `2` |
| All other errors (API HTTP classes, date validation, transport, unknown) | `1` |

Implementation SHOULD use `errors.As` on `*api.ClassifiedError` (and classify ad-hoc if raw `*api.HTTPError` reaches CLI).

---

## Process boundary

```text
cmd/singctl/main.go:
  if err := cli.Execute(); err != nil {
      os.Exit(cli.ExitCode(err))
  }
```

MUST NOT always `os.Exit(1)`.

---

## stderr presentation

`cli.Execute` continues to print errors to stderr. Prefer `ClassifiedError.Error()` / Message so users see catalog strings (e.g. `Error: invalid token…`), not only `HTTP 401`.

Wrapping with extra prefixes MAY exist but MUST NOT strip catalog text and MUST NOT leak tokens.

---

## `singctl config validate` (F05 update)

| Scenario | Exit | stderr / outcome |
|----------|------|------------------|
| No token | 2 | set-token hint; no network |
| Mock 2xx | 0 | remote success message |
| Mock 401 | 1 | contains invalid token / set-token catalog text |
| Mock 404 | 3 | entity not found |
| Mock 429 × N (exhausted) | 1 | `rate limited` catalog text; mock saw 3 requests |
| Mock 5xx | 1 | server error catalog text |
| Transport failure (optional unit) | 1 | reachability error; not exit 2 |

Retry: validate probe goes through session client with 429 RoundTripper.

---

## Testing notes

- Unit-test `ExitCode` table without full process if possible
- CLI tests via `executeForTest` + assert error Kind/message; process exit code via thin test of `ExitCode` or integration harness
- No new CRUD commands

---

## Relation to F07

F07 may expand pipe/stdin/stdout rules. F05 locks **numeric** exit mapping for taxonomy errors used by validate and future commands.
