# Contract: Public Exit Codes (F07)

**Feature**: `007-scriptability-exits` | **Date**: 2026-07-16
**Runtime SoT**: `cli.ExitCode` — [F05 cli-exit-codes.md](../../005-error-retry/contracts/cli-exit-codes.md)
**User docs**: `docs/scriptability.md` (создаётся при implement)

---

## Table (MUST match TZ §10)

| Code | Meaning (RU docs) | Meaning (short EN) |
|------|-------------------|--------------------|
| `0` | Успех | success |
| `1` | Ошибка API, операции, транспорта или использования CLI | API / operation / usage error |
| `2` | Ошибка конфигурации | configuration error |
| `3` | Сущность не найдена | not found |

---

## Mapping (unchanged from F05 + clarify)

| Condition | Code |
|-----------|------|
| `err == nil` | `0` |
| Kind NotFound | `3` |
| Kind Config | `2` |
| Misuse (unknown command/flag, invalid flag value e.g. `--output`) | `1` |
| All other errors | `1` |

---

## Process boundary

```text
cmd/singctl/main.go:
  if err := cli.Execute(); err != nil {
      os.Exit(cli.ExitCode(err))
  }
```

---

## Documentation surfaces

| Surface | Requirement |
|---------|-------------|
| `docs/scriptability.md` | Full table + short stream/pipe overview |
| `docs/README.md` | Link to scriptability.md |
| Root `singctl --help` | Brief mention of codes `0`/`1`/`2`/`3` + pointer to docs |

---

## DoD tests

- `config validate`: success → ExitCode 0; no token → 2; mock 404 → 3; mock 401/5xx/429 exhaust → 1
- Misuse: `--unknown-flag` or invalid `--output` → ExitCode `1` (not 2/3)
- `--help` stdout contains all four code numerals (or explicit exit-code blurb)
