# Quickstart: Scriptability & Exit Codes (F07)

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [exit-codes-public.md](./contracts/exit-codes-public.md), [stream-separation.md](./contracts/stream-separation.md), [pipe-scenarios.md](./contracts/pipe-scenarios.md).

---

## Prerequisites

```bash
# from repo root
make test
```

Токены в тестах — только `test-token-…` / `fake-…` (constitution VII). Live API не обязателен.

---

## Automated (primary DoD)

```bash
make test
```

Ожидаемо покрыто:

| Check | Maps to |
|-------|---------|
| `ExitCode` / `config validate` matrix 0/1/2/3 | SC-002, FR-001 |
| Misuse flag → exit 1, stdout empty | SC-002, FR-001a, FR-004 |
| `config show` success: data stdout, stderr empty | SC-003, FR-003 |
| Validate/show errors: stderr message, stdout empty | SC-003, FR-004 |
| `--help` mentions exit codes 0–3 | SC-001, FR-002 |
| `internal/output` pipe/ANSI / json-csv fixtures | SC-005, FR-008/009 |
| Docs file `docs/scriptability.md` present (test or checklist) | SC-001 |

---

## Manual spot-checks (optional)

```bash
go build -o /tmp/singctl ./cmd/singctl

/tmp/singctl --help | head -40
# expect: brief exit codes 0/1/2/3 + pointer to docs/scriptability.md

# misuse → 1
/tmp/singctl --unknown-flag; echo exit:$?
# expect: stderr message, exit 1

# streams (needs local config; use test-token-… only)
/tmp/singctl config show >/tmp/out.txt 2>/tmp/err.txt; echo exit:$?
# success: /tmp/err.txt empty; /tmp/out.txt has config
```

Сверить `docs/scriptability.md` с таблицей ТЗ §10 и [exit-codes-public.md](./contracts/exit-codes-public.md).

---

## Pipe scenarios (contract-level)

Полный live `task`/`time` — **F08+**. В F07 достаточно:

1. Матрица в [pipe-scenarios.md](./contracts/pipe-scenarios.md) (все 4 примера со статусом).
2. F06 unit: JSON array / CSV header / no ANSI non-TTY.
3. F07 stream/exit harness на `config *`.

---

## DoD checklist

- [x] `docs/scriptability.md` + ссылка в `docs/README.md`
- [x] Root `--help` mentions 0/1/2/3
- [x] Misuse → exit 1, empty stdout
- [x] `config show` / `config validate` stream rules
- [x] `make test` green; coverage not regress
- [x] Pipe §10 matrix complete (SC-004)
- [x] No new user-facing CLI commands
