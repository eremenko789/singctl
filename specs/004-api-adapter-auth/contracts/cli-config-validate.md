# Contract: singctl config validate (F04 update)

**Feature**: `004-api-adapter-auth` | **Date**: 2026-07-16
**Extends**: `specs/002-config-token/contracts/cli-config.md` (секция `validate`)

---

## Invocation

```text
singctl config validate
```

Глобальные флаги без изменений: `--config`, `--token`, …

---

## Behavior (F04)

### No token

- Exit ≠ 0
- Message hints `config set-token` (Russian text as in F02)
- No network request (factory fail-fast)

### Token present (file or `--token`)

1. Build API session from effective settings via `internal/api`
2. Run connectivity probe (`GET /v2/project` through adapter)
3. Outcomes:

| Probe result | Exit | User-visible message |
|--------------|------|----------------------|
| 2xx success | 0 | Confirms **remote** API check succeeded (MUST NOT say validation is only local stub / «заглушка») |
| HTTP non-2xx | ≠ 0 | Error; MUST NOT claim remote OK |
| Transport / timeout | ≠ 0 | Error; MUST NOT claim remote OK |
| Factory error (bad URL/timeout) | ≠ 0 | Error; no false OK |

### Secrets

- Full token MUST NOT appear on stdout/stderr

---

## Testing notes

- Prefer `httptest` + temp config with `api.base_url` = mock URL and `api.token` = `test-token-…`
- Replace F02 stub assertion («локаль»/«заглуш») with remote-success wording assertions
- Keep «no token → set-token hint» test from F02

---

## Out of scope for this command

- CRUD entity operations
- Retry on 429/5xx (F05)
- Interactive wizard
