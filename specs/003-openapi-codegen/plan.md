# Implementation Plan: OpenAPI Codegen Pipeline (F03)

**Branch**: `003-openapi-codegen` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-openapi-codegen/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F03. Он фиксирует дизайн/контракты и ожидаемую структуру реализации. Код и полные тест-сьюты создаются на фазе `/speckit-tasks` и `/speckit-implement`.

---
## Summary

F03 доводит OpenAPI-пайплайн до рабочего codegen:

- конфиг `api/oapi-codegen.yaml` (package `apiclient`, models+client → `internal/apiclient/client.gen.go`);
- Make-таргеты `openapi-fetch`, `api-coverage-check`, `generate` как **независимые** точки входа (рекомендуемый порядок только в docs/`make help`);
- закоммиченный снимок OpenAPI, матрица `docs/api/coverage.md`, конфиг и `*.gen.go` как DoD;
- `api-coverage-check`: JSON ops == `EXPECTED_API_OPS` (51) + наличие `coverage.md` (без парсинга строк матрицы).

Ручные HTTP-обёртки, CLI CRUD, auth/adapter (F04) и drift/no-diff gate — вне scope.

---
## Technical Context

**Language/Version**: Go 1.23 (модуль `github.com/eremenko789/singctl`); вспомогательный Python 3 только в Make для подсчёта operations.

**Primary Dependencies**: `oapi-codegen` v2 CLI (`go install …/oapi-codegen/v2/cmd/oapi-codegen@latest`); runtime-зависимости генератора подтягиваются в `go.mod` после первой успешной генерации / `go mod tidy`.

**Storage**: файлы в git — `docs/api/openapi.{json,yaml}`, `docs/api/coverage.md`, `api/oapi-codegen.yaml`, `internal/apiclient/*.gen.go`; `.env` не коммитится.

**Testing**: `make api-coverage-check` / `make generate` как контрактные проверки; `go build`/`go test` на пакете/модуле после появления gen (с исключением `*.gen.go` из coverage gate при настройке `make test`, constitution IX); TDD на новый **ручной** production-код, если он появится (скрипты/хелперы) — иначе достаточно контрактных Make-прогонов.

**Target Platform**: macOS/Linux разработчик + CI с `make`, `curl`, `python3`, `oapi-codegen` в PATH.

**Project Type**: developer tooling / codegen pipeline внутри Go CLI monorepo (не runtime CLI UX).

**Performance Goals**: N/A (локальные one-shot таргеты; generate < ~1 мин на полном снимке — ориентир, не SLA).

**Constraints**: constitution III/VIII/IX; независимые Make-таргеты; без drift-gate; без ручного HTTP CRUD; публичный upstream snapshot без токена.

**Scale/Scope**: 51 HTTP operations API v2; один сгенерированный файл клиента; три Make-таргета + конфиг + docs sync.

---
## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | gen-клиент — библиотечный пакет; не меняет модель поставки `singctl` |
| G3 | OpenAPI-Generated API Client | PASS | ядро F03: только codegen, без ручных HTTP CRUD |
| G4 | Shared Client for CLI and TUI | PASS (N/A) | адаптер/CLI/TUI — F04+; пакет общий задел |
| G5 | Scriptability First | PASS (N/A) | Make exit codes; не CLI пользователя |
| G6 | Honest API Boundaries | PASS | снимок/ops из реального OpenAPI; без обещаний сверх API |
| G7 | Security of Credentials | PASS | fetch публичный; `.env`/токены не в git; smoke вне DoD F03 |
| G8 | Makefile + `.env` | PASS | канон запуска — Make; параметры из `.env` / defaults |
| G9 | TDD & Coverage | PASS | gen MAY вне coverage; ручной код — TDD; `make test` не регрессирует |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---
## Project Structure

### Documentation (this feature)

```text
specs/003-openapi-codegen/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── make-openapi.md
└── tasks.md             # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
api/
└── oapi-codegen.yaml          # NEW (F03)

docs/api/
├── openapi.json               # exists; обновляется openapi-fetch
├── openapi.yaml               # exists; вход generate
└── coverage.md                # exists; наличие проверяет api-coverage-check

internal/apiclient/
└── client.gen.go              # NEW via make generate; commit (DoD)

Makefile                       # openapi-fetch / api-coverage-check / generate (уточнить атомарность fetch)
.env.example                   # уже есть; при необходимости комментарии к codegen
docs/openapi-codegen.md        # sync с фактическими таргетами
docs/makefile.md               # sync
```

**Structure Decision**: Single Go module. Codegen output только в `internal/apiclient/` (не `internal/api` — адаптер F04). Конфиг генератора в `api/oapi-codegen.yaml` по ТЗ/docs. CLI/entity-команды не трогаем.

---
## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---
## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
- формат конфига oapi-codegen и single-file output;
- независимые Make-таргеты и источники JSON vs YAML;
- строгость coverage-check;
- атомарность `openapi-fetch` (temp → rename);
- версия генератора (`@latest` v2) и коммит gen как source of truth;
- тестирование / coverage exclusion для `*.gen.go`.

---
## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — сущности снимка, конфига, ожидания coverage, артефакты gen
- `contracts/make-openapi.md` — контракт Make-таргетов (входы/выходы/exit codes)
- `quickstart.md` — офлайн/онлайн проверка пайплайна

---
## Next Phase (not executed here)

`/speckit-tasks` должен разложить:
1. тесты/контрактные проверки coverage-check и generate (TDD там, где появится ручной код);
2. `api/oapi-codegen.yaml` + успешный `make generate` + коммит `*.gen.go`;
3. доработка Makefile (атомарный fetch при необходимости) и sync docs;
4. исключение gen из coverage gate при касании `make test`.
