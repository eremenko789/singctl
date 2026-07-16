# Implementation Plan: API Adapter & Auth (F04)

**Branch**: `004-api-adapter-auth` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-api-adapter-auth/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F04. Он фиксирует дизайн/контракты и ожидаемую структуру реализации. Код и полные тест-сьюты создаются на фазе `/speckit-tasks` и `/speckit-implement`.

---

## Summary

F04 добавляет тонкий адаптер `internal/api/` поверх codegen (`internal/apiclient/`):

- фабрика сеанса из effective config (`base_url`, «голый» `token`, `timeout`) с fail-fast на пустой токен/URL;
- Bearer через `RequestEditor` (без двойного префикса);
- timeout через `http.Client.Timeout`;
- общий маппинг не-2xx → типизированная `HTTPError` (status + body) без retry/taxonomy (F05);
- happy path / `config validate`: один лёгкий вызов `ProjectController_list` (`GET /v2/project`) через `ClientWithResponses`.

Полные фасады сущностей и CLI CRUD — вне scope.

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (oapi-codegen v2 `ClientWithResponses`); `net/http` + `httptest` в тестах; `internal/config` (F02); `cobra` CLI для обновления `config validate`.

**Storage**: N/A (сеанс в памяти; конфиг пользователя уже в F02).

**Testing**: TDD — сначала failing unit-тесты `internal/api` (мок HTTP), затем фабрика/маппинг; затем обновление `config validate` + тесты CLI с `httptest` и `api.base_url` = URL мока. `make test` с coverage; фикстуры токенов `test-token-…` (constitution VII).

**Target Platform**: CLI `singctl` на macOS/Linux (как F01/F02).

**Project Type**: library adapter + тонкая CLI-интеграция (`config validate`) внутри Go CLI monorepo.

**Performance Goals**: N/A (локальные запросы; timeout из конфига, default 30s).

**Constraints**: constitution III/IV/VII/IX; без ручного CRUD/DTO; без retry; без полного фасада 51 ops; base URL — origin **без** суффикса `/v2` (paths codegen уже содержат `/v2/...`).

**Scale/Scope**: один пакет адаптера + правка `config validate`; одна репрезентативная операция для DoD/validate.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go-пакеты в том же модуле |
| G3 | OpenAPI-Generated API Client | PASS | вызовы только через `apiclient`; адаптер = auth/timeout/map |
| G4 | Shared Client for CLI and TUI | PASS | `internal/api` — каноническая фабрика сеанса для CLI/TUI |
| G5 | Scriptability First | PASS | `validate` exit codes; без ложного OK |
| G6 | Honest API Boundaries | PASS | один list-вызов; без обещаний сверх REST |
| G7 | Security of Credentials | PASS | fail-fast; фиктивные токены в тестах; без утечки в сообщениях |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` уже канон |
| G9 | TDD & Coverage | PASS | тесты до/вместе с кодом; ручной код в coverage gate |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/004-api-adapter-auth/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-session.md
│   └── cli-config-validate.md
└── tasks.md             # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/                 # NEW
├── session.go                # NewSession / NewFromSettings, Bearer editor, timeout HTTP client
├── errors.go                 # HTTPError (+ StatusCode), EnsureSuccess / MapResponse helpers
├── validate.go               # ValidateConnectivity → ProjectControllerListWithResponse
├── session_test.go
├── errors_test.go
└── validate_test.go

internal/apiclient/
└── client.gen.go             # EXISTS (F03); consume only

internal/config/              # EXISTS (F02); EffectiveSettings / Document inputs

internal/cli/
└── config_validate.go        # UPDATE: remote check via api.Session
└── config_validate_test.go   # UPDATE: httptest + stub message removed
```

**Structure Decision**: Single Go module. Адаптер только в `internal/api/` (не размазывать auth по `cli`). Codegen остаётся в `internal/apiclient/`. CLI вызывает фабрику сеанса и `ValidateConnectivity` (или эквивалент), не собирает HTTP вручную.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
- `ClientWithResponses` + `WithRequestEditorFn` + `WithHTTPClient`;
- нормализация Bearer и base URL (без `/v2`);
- timeout через `http.Client.Timeout`;
- типизированная `HTTPError` для F05;
- canonical probe op: `ProjectController_list`;
- тестирование через `httptest` без DI-фреймворка.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — Session, HTTPError, settings inputs
- `contracts/api-session.md` — контракт фабрики/маппинга/probe
- `contracts/cli-config-validate.md` — обновлённое поведение `config validate`
- `quickstart.md` — офлайн проверка `make test` + ручной validate против мока/API

---

## Next Phase (not executed here)

`/speckit-tasks` должен разложить (TDD):
1. тесты фабрики (fail-fast token/URL, Bearer, timeout) → реализация session;
2. тесты `HTTPError` / EnsureSuccess → errors;
3. happy path + non-2xx (0 retry) на `ProjectControllerList` → validate helper;
4. обновить `config validate` + CLI-тесты (убрать stub-assert);
5. godoc без package-stutter; coverage через `make test`.
