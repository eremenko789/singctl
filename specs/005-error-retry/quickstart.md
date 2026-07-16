# Quickstart: Error Handling & Retry (F05)

**Feature**: `005-error-retry` | **Date**: 2026-07-16

Офлайн-проверка DoD после `/speckit-implement`. Контракты: [api-errors-retry.md](./contracts/api-errors-retry.md), [cli-exit-codes.md](./contracts/cli-exit-codes.md).

---

## Prerequisites

- F04 merged/available (`internal/api` session + `HTTPError` + `config validate` remote probe)
- Go toolchain; `make test`

---

## 1. Unit + CLI tests

```bash
make test
```

**Expected**:
- exit 0
- taxonomy table tests green (401/403/404/422/429/5xx messages)
- retry tests: 429→success, 429 exhaust (3 requests), no retry on 404; sleeper-injected (suite still fast)
- `ParseDate` rejects bad input with `Expected: YYYY-MM-DD`
- `ExitCode` mapping 0/1/2/3
- `config validate` tests assert classified messages + exit semantics

---

## 2. Manual validate against mock (optional)

Поднимите mock, отвечающий 401 / 429×3 / 200 на `GET /v2/project`, укажите URL в конфиге:

```bash
singctl config validate
echo $?
```

**Expected examples**:
- 200 → `0` + remote OK
- 401 → `1` + invalid token message
- persistent 429 → `1` + `rate limited` (три попытки на стороне клиента)

---

## 3. No token

```bash
env XDG_CONFIG_HOME=/tmp/singctl-empty-$$ singctl config validate
echo $?
```

**Expected**: exit `2`, set-token hint, no network.

---

## Definition of done (checklist)

- [x] Classify + message catalog for 401/403/404/422/429/5xx
- [x] 429 retry (max 3) with exponential + `Retry-After` cap 30s
- [x] ParseDate unit tests (no date CLI command)
- [x] `cli.ExitCode` + `main` mapping; transport → 1
- [x] `config validate` wired to taxonomy/retry
- [x] No TUI banners; no entity CRUD
- [x] `make test` green; gitleaks-safe fixtures
