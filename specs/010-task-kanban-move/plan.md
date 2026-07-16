# Implementation Plan: Task Kanban Link & Move (F10)

**Branch**: `010-task-kanban-move` | **Date**: 2026-07-17 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/010-task-kanban-move/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F10. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F10 — CRUD связи задача↔канбан-колонка и UX перемещения: **`singctl task kanban`** (`list` / `get` / `create` / `update` / `delete`) + **`singctl task move`** поверх codegen KanbanTaskStatusController_* через тонкий фасад в `internal/api`, рендер F06 (SingleObject из F08) и exit/stream контракты F05/F07.

Ключевые решения clarify + research:
- list: опциональные `--task` / `--status`; без `--limit` / `--offset` / `--removed`; **без** pre-check задачи/колонки;
- create: `--task` + `--column`, опциональный `--order`; pre-check GetTask; **без** клиентской уникальности (дубликаты допускаем);
- update: частичный `--task` / `--column` / `--order`; без pre-check;
- move: GetTask → list по taskId → 0 create / 1 update statusId (даже same column) / >1 exit `1`; **без** `--order`;
- stdout: create/update/get/move → полная связь; delete → пусто; json/yaml list → массив, одна запись → объект;
- F13 (колонки) и TUI move — out of scope.

DoD: все 5 KanbanTaskStatusController_* через фасад + unit-тесты мок-HTTP; ветвление move 0/1/>1; CLI help; CLI harness на httptest.

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (codegen KanbanTaskStatusController_*), `internal/api` (Session, GetTask, Classify, EnsureSuccess), `internal/cli` (cobra, ExitCode, Opts, task group), `internal/output` (Render + SingleObject), `internal/config`. Без новых внешних библиотек.

**Storage**: N/A (состояние на стороне SingularityApp API).

**Testing**: TDD — сначала failing тесты фасада (`internal/api/kanban_task_status_test.go` или `kanban.go` + tests) и CLI (`task_kanban_*_test.go`, `task_move_test.go`, help); httptest + `test-token-…`; `executeForTest` + `ExitCode`; `make test` + coverage gate.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: entity CLI + API facade в Go monorepo (слой B, зависит от F08; F09 соседний, не блокирует).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD (live — F33). Move = GetTask + list + create|update (2–3 round-trip осознанно).

**Constraints**: constitution III (только codegen HTTP), IV (фасад общий для будущего TUI; логика move в api), V/F07 streams+exits, VI (нет column CRUD / TUI / pagination flags / `--order` на move), VII (фикстуры токенов), IX (TDD + coverage).

**Scale/Scope**: 5 подкоманд `task kanban` + 1 `task move`; 5 API operations + оркестрация Move; минимальный набор колонок link view; reuse SingleObject.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS | фасад вызывает `KanbanTaskStatusController*WithResponse`; без ручных HTTP DTO |
| G4 | Shared Client for CLI and TUI | PASS | CRUD + `MoveTaskToKanban` в `internal/api`; CLI: validate → GetTask (create/move) → facade → render |
| G5 | Scriptability First | PASS | F06/F07 + clarify stdout/json shape / move ambiguous / exits |
| G6 | Honest API Boundaries | PASS | нет column CRUD; нет `--order` на move; list без pagination flags; create не upsert |
| G7 | Security of Credentials | PASS | `test-token-…`; токен не в stdout |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты фасада и CLI до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/010-task-kanban-move/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-kanban-facade.md    # Session methods → KanbanTaskStatusController_* + Move
│   ├── cli-kanban.md           # task kanban * + task move, flags, exits
│   └── kanban-output.md        # колонки RecordSet; list array vs single object
└── tasks.md                    # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── kanban_task_status.go       # NEW: List/Get/Create/Update/Delete + MoveTaskToKanban
├── kanban_task_status_test.go  # NEW: httptest per op + move 0/1/>1 + not-found
├── task.go                     # EXISTS: GetTask reused for CLI pre-check
└── doc.go                      # UPDATE: упомянуть kanban-task-status facade

internal/cli/
├── task_cmd.go                 # UPDATE: AddCommand kanban + move; help Long
├── task_kanban_cmd.go          # NEW: группа kanban
├── task_kanban_list.go         # NEW
├── task_kanban_get.go          # NEW
├── task_kanban_create.go       # NEW
├── task_kanban_update.go       # NEW
├── task_kanban_delete.go       # NEW
├── task_move.go                # NEW: task move
├── task_kanban_render.go       # NEW: KanbanLink → RecordSet / columns
├── task_kanban_*_test.go       # NEW: help, validate, exit matrix
├── task_move_test.go           # NEW: move branches + pre-check
└── (reuse) exit.go, output.go, openAPISession, executeForTest

internal/output/                # EXISTS SingleObject — reuse
internal/apiclient/             # EXISTS codegen — DO NOT hand-edit
docs/api/coverage.md            # OPTIONAL later: отметить закрытие KanbanTaskStatus_* (F35)
```

**Structure Decision**: Фасад kanban-task-status + оркестрация `MoveTaskToKanban` в `internal/api` (Session + Classify). Pre-check задачи — в **CLI** через `GetTask` (create/move). List без GetTask. Рендер через `internal/output` SingleObject. Подгруппа `task kanban` и команда `task move` регистрируются из `task_cmd.go`.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
1. Фасад API: методы на `*Session` + `MoveTaskToKanban`.
2. Pre-check GetTask только в CLI (create/move); list без pre-check.
3. List params: только TaskId/StatusId при наличии флагов.
4. Create без клиентской уникальности; move >1 → KindValidation / sentinel → exit 1.
5. Move: always update statusId when exactly one link (incl. same column); no kanbanOrder on move.
6. Колонки вывода; валидация CLI; тестовый паттерн httptest.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — KanbanLink view, ListQuery, WriteInput, MoveIntent
- `contracts/api-kanban-facade.md` — сигнатуры и mapping
- `contracts/cli-kanban.md` — CLI surface + validation + exits
- `contracts/kanban-output.md` — RecordSet / SingleObject
- `quickstart.md` — `make test` + ручной help smoke

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. api kanban facade tests (5 ops + 404) → `kanban_task_status.go`;
2. MoveTaskToKanban tests (0/1/>1) → facade method;
3. CLI help registration → `task_kanban_cmd` + `task_move` + `task_cmd` update;
4. list/get/create/update/delete CLI tests → commands + render;
5. move CLI (pre-check, branches) + stream/exit matrix; `make test` coverage.
