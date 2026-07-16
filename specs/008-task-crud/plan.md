# Implementation Plan: Task CRUD (F08)

**Branch**: `008-task-crud` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/008-task-crud/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F08. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F08 — первая entity-группа CLI: **`singctl task`** (`list`/`get`/`create`/`update`/`delete`/`archive`/`trash`) поверх codegen TaskController_* через тонкий фасад в `internal/api`, рендер F06 и exit/stream контракты F05/F07.

Ключевые решения clarify + research:
- create/update: флаги ТЗ §6.1 **плюс** `--project` / `--parent`; `--note` as-is;
- stdout: create/update/archive/trash → полная задача; delete → пусто;
- json/yaml: list → массив; одна задача → **один объект** (расширение `internal/output`);
- `--limit` ∈ 1…1000, `--offset` ≥ 0 — валидация CLI до сети;
- `--delete-date` на create: OpenAPI create DTO **без** `deleteDate` → POST create, затем PATCH update (честная граница API).

DoD: все 5 TaskController_* через фасад + unit-тесты мок-HTTP; CLI help; CLI harness на httptest (как `config validate`).

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (codegen TaskController_*), `internal/api` (Session, Classify, ParseDate, retry), `internal/cli` (cobra, ExitCode, Opts), `internal/output` (Render + SingleObject), `internal/config` (LoadEffectiveSettings). Без новых внешних библиотек.

**Storage**: N/A (состояние на стороне SingularityApp API).

**Testing**: TDD — сначала failing тесты фасада (`internal/api/task_test.go`) и CLI (`task_*_test.go` / help); httptest + `test-token-…`; `executeForTest` + `ExitCode`; `make test` + coverage gate.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: entity CLI + API facade в Go monorepo (первая из слоя B).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD (live — F33).

**Constraints**: constitution III (только codegen HTTP), IV (фасад общий для будущего TUI), V/F07 streams+exits, VI (нет checklist/kanban/`move`; note без fake-delta), VII (фикстуры токенов), IX (TDD + coverage).

**Scale/Scope**: 7 подкоманд; 5 API operations; минимальный набор колонок task; точечное расширение output для single-object json/yaml.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS | фасад вызывает `*WithResponse`; без ручных HTTP DTO |
| G4 | Shared Client for CLI and TUI | PASS | фасад в `internal/api`; CLI только flags→facade→render |
| G5 | Scriptability First | PASS | F06/F07 + clarify stdout/json shape/limit |
| G6 | Honest API Boundaries | PASS | `--delete-date` create = create+update; note delta честно в help; нет F09/F10 |
| G7 | Security of Credentials | PASS | `test-token-…`; токен не в stdout |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты фасада и CLI до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/008-task-crud/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-task-facade.md     # Session methods → TaskController_*
│   ├── cli-task.md            # команды, флаги, валидация, exit/streams
│   └── task-output.md         # колонки RecordSet; list array vs single object
└── tasks.md                   # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── task.go                 # NEW: List/Get/Create/Update/Delete (+ Archive/Trash helpers)
├── task_test.go            # NEW: httptest per TaskController op + not-found
├── date.go                 # EXISTS: ParseDate; MAY add TodayCalendarDate helper
└── doc.go                  # UPDATE: упомянуть task facade

internal/output/
├── model.go / render_*.go  # UPDATE: SingleObject (или RenderOne) для json/yaml одного объекта
└── *_test.go               # UPDATE: single-object + list array regression

internal/cli/
├── root.go                 # UPDATE: AddCommand(newTaskCmd())
├── task_cmd.go             # NEW: группа task + registration
├── task_list.go            # NEW
├── task_get.go             # NEW
├── task_create.go          # NEW
├── task_update.go          # NEW
├── task_delete.go          # NEW
├── task_archive.go         # NEW
├── task_trash.go           # NEW
├── task_render.go          # NEW: Task → RecordSet / columns
├── task_*_test.go          # NEW: help, filters, mutate stdout, exit matrix
└── (reuse) exit.go, output.go, executeForTest harness

internal/apiclient/         # EXISTS codegen — DO NOT hand-edit
docs/api/coverage.md        # OPTIONAL later: отметить закрытие TaskController_* (F35 gate)
```

**Structure Decision**: Фасад задач в `internal/api` (переиспользуем Session/Classify/EnsureSuccess). CLI — тонкие cobra-команды. Рендер через `internal/output` с минимальным расширением SingleObject. Archive/trash — helpers фасада или CLI→Update с одной датой; delete — Delete + пустой stdout.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
1. Фасад API: методы на `*Session`, маппинг в codegen DTO, Classify+WithEntityID.
2. `--delete-date` на create → POST затем PATCH.
3. «Сегодня» для archive/trash → локальный календарный `YYYY-MM-DD`.
4. Single-object json/yaml → расширение `internal/output`.
5. Колонки list/get (минимальный стабильный набор).
6. Валидация limit/offset/priority/ID/title/update-empty на CLI.
7. Паттерн тестов: httptest + executeForTest (как validate).

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — Task view, ListQuery, WriteInput, Archive/Trash intent
- `contracts/api-task-facade.md` — сигнатуры и mapping
- `contracts/cli-task.md` — CLI surface + validation + exits
- `contracts/task-output.md` — RecordSet / SingleObject
- `quickstart.md` — `make test` + ручной help smoke

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. output SingleObject tests → реализация;
2. api task facade tests (5 ops + 404) → `task.go`;
3. CLI help registration tests → `task_cmd` + root;
4. list/get tests (filters, json shape, empty) → commands + render;
5. create/update validation + happy path → commands;
6. archive/trash/delete → commands;
7. stream/exit matrix на task командах; `make test` coverage.
