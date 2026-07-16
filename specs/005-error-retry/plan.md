# Implementation Plan: Error Handling & Retry (F05)

**Branch**: `005-error-retry` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-error-retry/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F05. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F05 наращивает `internal/api` поверх контракта F04 (`HTTPError` + session):

- taxonomy HTTP → стабильные user-facing сообщения (ТЗ §8.1) и класс ошибки;
- automatic retry только для **429**: макс. **3** HTTP-запроса, пауза = `Retry-After` (с потолком **30s**) или exponential (**1s**, **2s**);
- клиентские ошибки: унификация «нет токена» + `ParseDate` (`Expected: YYYY-MM-DD`) без CLI-команды с датой;
- CLI: общий `ExitCode(err)` (0/1/2/3) + проводка в `config validate` и `main` (транспорт → 1).

TUI-баннеры и CRUD-команды — вне scope.

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/api` / `internal/apiclient` (F04/F03); `net/http` + `httptest`; `internal/cli` (cobra); без новых внешних retry-библиотек.

**Storage**: N/A.

**Testing**: TDD — сначала failing unit-тесты taxonomy/retry/date/exit; затем реализация; затем обновление `config validate` + `main` exit mapping. Backoff в тестах — injectable sleeper (нулевые/мгновенные паузы). `make test` + coverage; токены `test-token-…`.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: library adapter extensions + тонкая CLI-проводка внутри Go monorepo.

**Performance Goals**: unit-suite быстрый (без реальных sleep 1s/2s в CI); production backoff — секунды, не минуты (потолок `Retry-After` 30s).

**Constraints**: constitution III/IV/V/VII/IX; retry только 429; без ручного CRUD; без TUI; сообщения taxonomy — стабильные EN-строки из ТЗ/clarify.

**Scale/Scope**: расширение `internal/api` + `internal/cli` exit helper + правка `cmd/singctl/main.go` и `config validate`; без новых entity commands.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go в том же модуле |
| G3 | OpenAPI-Generated API Client | PASS | retry через HTTP transport поверх codegen; без ручных DTO CRUD |
| G4 | Shared Client for CLI and TUI | PASS | taxonomy/retry в `internal/api`; CLI только ExitCode + validate |
| G5 | Scriptability First | PASS | exit 0/1/2/3; стабильные тексты ошибок |
| G6 | Honest API Boundaries | PASS | без новых API-обещаний |
| G7 | Security of Credentials | PASS | токен не в сообщениях; фиктивные фикстуры |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/005-error-retry/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-errors-retry.md
│   └── cli-exit-codes.md
└── tasks.md             # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── errors.go             # EXISTS: HTTPError, EnsureSuccess — UPDATE: Classify → ClassifiedError
├── retry.go              # NEW: RoundTripper / policy (429, Retry-After, exponential)
├── date.go               # NEW: ParseDate + DateError
├── errors_test.go        # UPDATE/extend taxonomy table
├── retry_test.go         # NEW
├── date_test.go          # NEW
├── session.go            # UPDATE: wire retry transport into HTTP client
├── validate.go           # UPDATE: Classify on probe failures; retry via transport
└── …

internal/cli/
├── exit.go               # NEW: ExitCode(err) → 0/1/2/3
├── exit_test.go          # NEW
├── config_validate.go    # UPDATE: surface Classified messages; rely on ExitCode
├── config_validate_test.go
└── root.go / Execute     # UPDATE: return/propagate so main can map codes (if needed)

cmd/singctl/
└── main.go               # UPDATE: os.Exit(cli.ExitCode(err)) instead of always 1
```

**Structure Decision**: Taxonomy + retry + ParseDate живут в `internal/api` (общий слой для CLI/TUI). Числовые exit codes — тонкий helper в `internal/cli` + `main`, чтобы process boundary не тянуть в library-пакет с `os.Exit` внутри api.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения зафиксированы в `research.md`:
- retry via custom `http.RoundTripper` на session client;
- константы: 3 attempts, base 1s/2s, `Retry-After` cap 30s;
- `Classify(*HTTPError)` + entity ID option для 404;
- 422 body extract + fallback;
- ParseDate only + unit tests;
- ExitCode mapping including transport → 1.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — ClassifiedError, RetryPolicy, DateError, Exit semantics
- `contracts/api-errors-retry.md` — Classify / retry / ParseDate
- `contracts/cli-exit-codes.md` — ExitCode + validate wiring
- `quickstart.md` — `make test` + негативные сценарии validate

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. taxonomy Classify tests → implementation;
2. retry RoundTripper tests (429 success/exhaust/`Retry-After`/no-retry-other) → implementation + session wire;
3. ParseDate tests → date helper;
4. ExitCode tests → cli helper + main;
5. config validate tests (401/404/429 messages + exit codes) → wiring;
6. godoc; `make test` coverage.
