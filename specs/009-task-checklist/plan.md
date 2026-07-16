# Implementation Plan: Task Checklist (F09)

**Branch**: `009-task-checklist` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/009-task-checklist/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F09. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F09 — CRUD чек-листа задачи: **`singctl task checklist`** (`list` / `get` / `add` / `update` / `delete`) поверх codegen ChecklistItemController_* через тонкий фасад в `internal/api`, рендер F06 (SingleObject из F08) и exit/stream контракты F05/F07.

Ключевые решения clarify + research:
- list: только parent = `<TASK_ID>`; без `--limit` / `--offset` / `--removed`;
- без `--order` / `parentOrder` / смены parent на write;
- перед `list` / `add`: pre-check через `GetTask` (F08); not found → exit `3`, checklist API не вызывается; ответ задачи не в stdout;
- `--title`: обязателен на add; пустой/whitespace (после trim) → exit `1` до сети; на update — если флаг задан;
- `--done` / `--undone` на update (взаимоисключающие); на add — опциональный `--done`;
- stdout: add/update/get → полный пункт; delete → пусто; json/yaml list → массив, одна запись → объект.

DoD: все 5 ChecklistItemController_* через фасад + unit-тесты мок-HTTP; CLI help; CLI harness на httptest (pre-check + checklist).

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (codegen ChecklistItemController_*), `internal/api` (Session, GetTask, Classify, EnsureSuccess), `internal/cli` (cobra, ExitCode, Opts, task group), `internal/output` (Render + SingleObject из F08), `internal/config`. Без новых внешних библиотек.

**Storage**: N/A (состояние на стороне SingularityApp API).

**Testing**: TDD — сначала failing тесты фасада (`internal/api/checklist_test.go`) и CLI (`task_checklist_*_test.go` / help); httptest + `test-token-…`; `executeForTest` + `ExitCode`; `make test` + coverage gate.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: entity CLI + API facade в Go monorepo (слой B, зависит от F08).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD (live — F33). Pre-check добавляет один round-trip на list/add (осознанно).

**Constraints**: constitution III (только codegen HTTP), IV (фасад общий для будущего TUI), V/F07 streams+exits, VI (нет TUI checklist / kanban / `--order` / pagination flags), VII (фикстуры токенов), IX (TDD + coverage). Зависит от F08 `GetTask` и output SingleObject.

**Scale/Scope**: 5 подкоманд под `task checklist`; 5 API operations; минимальный набор колонок checklist item; без расширения output beyond F08 SingleObject.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS | фасад вызывает `ChecklistItemController*WithResponse`; без ручных HTTP DTO |
| G4 | Shared Client for CLI and TUI | PASS | фасад в `internal/api`; CLI: validate → GetTask (list/add) → facade → render |
| G5 | Scriptability First | PASS | F06/F07 + clarify stdout/json shape / empty title / pre-check exits |
| G6 | Honest API Boundaries | PASS | нет фейковых pagination/order; parent = task; TUI out of scope |
| G7 | Security of Credentials | PASS | `test-token-…`; токен не в stdout |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты фасада и CLI до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/009-task-checklist/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-checklist-facade.md   # Session methods → ChecklistItemController_*
│   ├── cli-checklist.md          # команды, флаги, pre-check, exit/streams
│   └── checklist-output.md       # колонки RecordSet; list array vs single object
└── tasks.md                      # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── checklist.go            # NEW: List/Get/Create/Update/Delete checklist items
├── checklist_test.go       # NEW: httptest per ChecklistItemController op + not-found
├── task.go                 # EXISTS: GetTask reused for CLI pre-check (no change required unless helper)
└── doc.go                  # UPDATE: упомянуть checklist facade

internal/cli/
├── task_cmd.go             # UPDATE: AddCommand(newTaskChecklistCmd()); help text
├── task_checklist_cmd.go   # NEW: группа checklist + registration
├── task_checklist_list.go  # NEW
├── task_checklist_get.go   # NEW
├── task_checklist_add.go   # NEW
├── task_checklist_update.go # NEW
├── task_checklist_delete.go # NEW
├── task_checklist_render.go # NEW: ChecklistItem → RecordSet / columns
├── task_checklist_*_test.go # NEW: help, pre-check, validate, exit matrix
└── (reuse) exit.go, output.go, openAPISession, executeForTest, task render helpers

internal/output/            # EXISTS SingleObject from F08 — reuse, no required change
internal/apiclient/         # EXISTS codegen — DO NOT hand-edit
docs/api/coverage.md        # OPTIONAL later: отметить закрытие ChecklistItemController_* (F35)
```

**Structure Decision**: Фасад чек-листа в `internal/api` (Session + Classify). Pre-check parent — в **CLI** через существующий `GetTask` (не смешивать task get в checklist facade). Рендер через `internal/output` SingleObject. Подгруппа `task checklist` регистрируется из `task_cmd.go`.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
1. Фасад API: методы на `*Session`, маппинг codegen DTO, Classify+WithEntityID.
2. Pre-check parent в CLI через `GetTask`, не в checklist facade.
3. List params: только `Parent`; остальные query unset.
4. `--done` / `--undone` → pointer bool в Update DTO.
5. Колонки list/get (id, title, done, parent, parentOrder).
6. Валидация title/ID/done+undone/update-empty на CLI.
7. Паттерн тестов: httptest (два endpoint для list/add) + executeForTest.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — ChecklistItem view, ListQuery, WriteInput
- `contracts/api-checklist-facade.md` — сигнатуры и mapping
- `contracts/cli-checklist.md` — CLI surface + pre-check + validation + exits
- `contracts/checklist-output.md` — RecordSet / SingleObject
- `quickstart.md` — `make test` + ручной help smoke

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. api checklist facade tests (5 ops + 404) → `checklist.go`;
2. CLI help registration tests → `task_checklist_cmd` + `task_cmd` update;
3. list/get tests (pre-check, json shape, empty) → commands + render;
4. add/update validation + happy path → commands;
5. delete + stream/exit matrix; `make test` coverage.
