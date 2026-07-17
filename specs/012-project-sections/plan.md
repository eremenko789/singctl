# Implementation Plan: Project Sections (F12)

**Branch**: `012-project-sections` | **Date**: 2026-07-17 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/012-project-sections/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F12. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F12 — CLI-подгруппа **`singctl project section`** (`list`/`get`/`create`/`update`/`delete`) поверх codegen TaskGroupController_* через тонкий фасад в `internal/api`, рендер F06 (`SingleObject`) и exit/stream контракты F05/F07. Паттерн — зеркало F11 (`project`) / F09 (`task checklist`), без канбан-колонок (F13).

Ключевые решения clarify + research:
- UX-термин **section**; API resource **task-group** / TaskGroup* — только в фасаде/contracts;
- list: обязательный `<PROJECT_ID>` → query `parent`; `--removed` / `--limit` / `--offset`;
- create: `<PROJECT_ID>` + непустой `--title` (trim); без флага `--parent`;
- update: `--title` и/или `--parent` (перенос); без `--order` / `externalId` / `fake`;
- пустой/whitespace `--title` или ID → exit `1` до сети;
- stdout: create/update → полная секция; delete → пусто; list → массив; get → один объект;
- `--limit` ∈ 1…1000, `--offset` ≥ 0 — валидация CLI до сети.

DoD: все 5 TaskGroupController_* через фасад + unit-тесты мок-HTTP; CLI help (группа + 5 команд); CLI harness на httptest (как project/checklist).

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (codegen TaskGroupController_*), `internal/api` (Session, Classify, EnsureSuccess, retry), `internal/cli` (cobra, ExitCode, Opts, существующая группа `project`), `internal/output` (Render + SingleObject), `internal/config`. Без новых внешних библиотек.

**Storage**: N/A (состояние на стороне SingularityApp API).

**Testing**: TDD — сначала failing тесты фасада (`internal/api/section_test.go`) и CLI (`project_section_*_test.go` / help); httptest + `test-token-…`; `executeForTest` + `ExitCode`; `make test` + coverage gate.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: entity CLI + API facade в Go monorepo (слой B, после project CRUD F11).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD (live — F33).

**Constraints**: constitution III (только codegen HTTP), IV (фасад общий для будущего TUI), V/F07 streams+exits, VI (нет column; нет fake «полного» OpenAPI write dump), VII (фикстуры токенов), IX (TDD + coverage). Depends F11.

**Scale/Scope**: 1 подгруппа + 5 подкоманд; 5 API operations; минимальный набор колонок section; без изменений `internal/output`; без archive/trash секций.

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS | фасад вызывает `TaskGroupController*WithResponse`; без ручных HTTP DTO |
| G4 | Shared Client for CLI and TUI | PASS | фасад в `internal/api`; CLI только flags→facade→render |
| G5 | Scriptability First | PASS | F06/F07 + clarify stdout/json shape/limit/required project id |
| G6 | Honest API Boundaries | PASS | нет column; нет order/externalId/fake; term section vs task-group честно в contracts |
| G7 | Security of Credentials | PASS | `test-token-…`; токен не в stdout |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты фасада и CLI до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/012-project-sections/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-section-facade.md   # Session methods → TaskGroupController_*
│   ├── cli-section.md          # команды, флаги, валидация, exit/streams
│   └── section-output.md       # колонки RecordSet; list array vs single object
└── tasks.md                    # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── section.go              # NEW: List/Get/Create/Update/Delete sections
├── section_test.go         # NEW: httptest per TaskGroupController op + not-found
└── doc.go                  # UPDATE: упомянуть section / TaskGroup facade

internal/cli/
├── project_cmd.go                # UPDATE: AddCommand(newProjectSectionCmd()); Long/help
├── project_section_cmd.go        # NEW: группа section + registration
├── project_section_list.go       # NEW
├── project_section_get.go        # NEW
├── project_section_create.go     # NEW
├── project_section_update.go     # NEW
├── project_section_delete.go     # NEW
├── project_section_render.go     # NEW: Section → RecordSet / columns
├── project_section_*_test.go     # NEW: help, filters, mutate stdout, exit matrix
└── (reuse) exit.go, output.go, executeForTest, project helpers

internal/output/            # EXISTS SingleObject — no change expected
internal/apiclient/         # EXISTS codegen — DO NOT hand-edit
docs/api/coverage.md        # OPTIONAL later: отметить закрытие TaskGroupController_* (F35)
```

**Structure Decision**: Фасад секций в `internal/api/section.go` (имена Section*; внутри TaskGroupController_*). CLI — `project section` как вложенная группа (зеркало `task checklist`). Рендер через существующий `SingleObject`.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
1. Фасад API: методы на `*Session`, unwrap list `taskGroups`, Classify+WithEntityID.
2. Именование Section vs TaskGroup.
3. Колонки list/get (минимальный стабильный набор).
4. Валидация limit/offset/ID/title/update-empty/parent на CLI.
5. Паттерн тестов: httptest `/v2/task-group` + executeForTest.
6. Регистрация `section` под `project`; help без column; обновить Long project.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — Section view, ListQuery, WriteInput
- `contracts/api-section-facade.md` — сигнатуры и mapping
- `contracts/cli-section.md` — CLI surface + validation + exits
- `contracts/section-output.md` — RecordSet / SingleObject
- `quickstart.md` — `make test` + ручной help smoke

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. api section facade tests (5 ops + 404) → `section.go`;
2. CLI help registration tests → `project_section_cmd` + project_cmd;
3. list/get tests (required project id, filters, json shape) → commands + render;
4. create/update validation + `--parent` move → commands;
5. delete + stream/exit matrix; `make test` coverage.
