# Implementation Plan: Project CRUD (F11)

**Branch**: `011-project-crud` | **Date**: 2026-07-17 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/011-project-crud/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F11. Код и полные тест-сьюты — на фазе `/speckit-tasks` / `/speckit-implement`.

---

## Summary

F11 — CLI-группа **`singctl project`** (`list`/`get`/`create`/`update`/`delete`/`archive`/`trash`) поверх codegen ProjectController_* через тонкий фасад в `internal/api`, рендер F06 (`SingleObject`) и exit/stream контракты F05/F07. Паттерн — зеркало F08 (task), без секций/колонок (F12/F13).

Ключевые решения clarify + research:
- create/update: ТЗ §6.2 (`title`, `note`, `notebook`, `emoji`, `color`) **плюс** `--parent`; без `--archive-date`/`--delete-date`;
- archive/trash: dedicated-команды → PATCH `journalDate` / `deleteDate`; default date = `TodayCalendarDate()`;
- `--emoji`: unicode → lowercase hex в CLI; hex 4–8 chars as-is; иначе exit `1` до сети;
- create response: unwrap `project` из `ProjectCreateResponseDto`; `taskGroup` игнорируется;
- stdout: create/update/archive/trash → полный проект; delete → пусто; list → массив; single → один объект;
- `--limit` ∈ 1…1000, `--offset` ≥ 0 — валидация CLI до сети;
- shared projects / section / column — не обещать (constitution VI).

DoD: все 5 ProjectController_* через фасад + unit-тесты мок-HTTP; CLI help (7 команд); CLI harness на httptest (как task).

---

## Technical Context

**Language/Version**: Go 1.24 (модуль `github.com/eremenko789/singctl`).

**Primary Dependencies**: `internal/apiclient` (codegen ProjectController_*), `internal/api` (Session, Classify, ParseDate, TodayCalendarDate, retry), `internal/cli` (cobra, ExitCode, Opts), `internal/output` (Render + SingleObject — уже есть), `internal/config`. Без новых внешних библиотек (emoji → hex на stdlib `unicode`/`utf8`/`strconv`).

**Storage**: N/A (состояние на стороне SingularityApp API).

**Testing**: TDD — сначала failing тесты фасада (`internal/api/project_test.go`) и CLI (`project_*_test.go` / help / emoji); httptest + `test-token-…`; `executeForTest` + `ExitCode`; `make test` + coverage gate.

**Target Platform**: CLI `singctl` (macOS/Linux).

**Project Type**: entity CLI + API facade в Go monorepo (слой B, после task/checklist/kanban-link).

**Performance Goals**: unit/CLI harness быстрый; без live API smoke в DoD (live — F33).

**Constraints**: constitution III (только codegen HTTP), IV (фасад общий для будущего TUI F20), V/F07 streams+exits, VI (нет section/column/shared; note без fake-delta), VII (фикстуры токенов), IX (TDD + coverage).

**Scale/Scope**: 7 подкоманд; 5 API operations; минимальный набор колонок project; emoji helper; без изменений `internal/output` (SingleObject reuse).

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | plan/research/data-model/contracts/quickstart по Spec Kit |
| G2 | Go Single Binary | PASS | только Go; один бинарь |
| G3 | OpenAPI-Generated API Client | PASS | фасад вызывает `ProjectController*WithResponse`; без ручных HTTP DTO |
| G4 | Shared Client for CLI and TUI | PASS | фасад в `internal/api`; CLI только flags→facade→render |
| G5 | Scriptability First | PASS | F06/F07 + clarify stdout/json shape/limit/archive |
| G6 | Honest API Boundaries | PASS | shared projects не обещать; create unwrap project only; нет F12/F13; note delta в help |
| G7 | Security of Credentials | PASS | `test-token-…`; токен не в stdout |
| G8 | Makefile + `.env` | PASS (N/A) | новых Make-таргетов не требуется; `make test` |
| G9 | TDD & Coverage | PASS | тесты фасада и CLI до/вместе с кодом; ручной код в coverage |

Post-design: все gates PASS / PASS N/A — Complexity Tracking пуст.

---

## Project Structure

### Documentation (this feature)

```text
specs/011-project-crud/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── api-project-facade.md   # Session methods → ProjectController_*
│   ├── cli-project.md          # команды, флаги, валидация, exit/streams
│   └── project-output.md       # колонки RecordSet; list array vs single object
└── tasks.md                    # /speckit-tasks — NOT created here
```

### Source Code (repository root)

```text
internal/api/
├── project.go              # NEW: List/Get/Create/Update/Delete (+ Archive/Trash)
├── project_test.go         # NEW: httptest per ProjectController op + not-found
├── emoji.go                # NEW (or in cli): NormalizeProjectEmoji hex/unicode
├── emoji_test.go           # NEW
├── date.go                 # EXISTS: ParseDate, TodayCalendarDate — reuse
└── doc.go                  # UPDATE: упомянуть project facade

internal/cli/
├── root.go                 # UPDATE: AddCommand(newProjectCmd())
├── project_cmd.go          # NEW: группа project + registration
├── project_list.go         # NEW
├── project_get.go          # NEW
├── project_create.go       # NEW
├── project_update.go       # NEW
├── project_delete.go       # NEW
├── project_archive.go      # NEW
├── project_trash.go        # NEW
├── project_render.go       # NEW: Project → RecordSet / columns
├── project_*_test.go       # NEW: help, filters, mutate stdout, emoji, exit matrix
└── (reuse) exit.go, output.go, executeForTest harness

internal/output/            # EXISTS SingleObject — no change expected
internal/apiclient/         # EXISTS codegen — DO NOT hand-edit
docs/api/coverage.md        # OPTIONAL later: отметить закрытие ProjectController_* (F35)
```

**Structure Decision**: Фасад проектов в `internal/api` (зеркало task). CLI — тонкие cobra-команды. Рендер через существующий `SingleObject`. Emoji-нормализация — чистая функция в `internal/api` (или `internal/cli`), вызываемая до фасада write, с unit-тестами. Archive/trash — helpers фасада над Update.

---

## Complexity Tracking

> Нет нарушений constitution — таблица пуста.

---

## Phase 0: Outline & Research (Output = research.md)

Решения в `research.md`:
1. Фасад API: методы на `*Session`, unwrap create `project`, Classify+WithEntityID.
2. Emoji: hex pass-through vs unicode→hex; multi-rune reject.
3. Archive/trash: TodayCalendarDate; без date flags на create/update.
4. Колонки list/get (минимальный стабильный набор).
5. Валидация limit/offset/ID/title/update-empty/emoji на CLI.
6. Паттерн тестов: httptest + executeForTest (как task).
7. Регистрация `project` в root; help без section/column.

---

## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md` — Project view, ListQuery, WriteInput, Archive/Trash intent
- `contracts/api-project-facade.md` — сигнатуры и mapping
- `contracts/cli-project.md` — CLI surface + validation + exits
- `contracts/project-output.md` — RecordSet / SingleObject
- `quickstart.md` — `make test` + ручной help smoke

---

## Next Phase (not executed here)

`/speckit-tasks` (TDD order):
1. api emoji normalize tests → helper;
2. api project facade tests (5 ops + 404 + archive/trash) → `project.go`;
3. CLI help registration tests → `project_cmd` + root;
4. list/get tests (filters, json shape, empty) → commands + render;
5. create/update validation + emoji + happy path → commands;
6. archive/trash/delete → commands;
7. stream/exit matrix; `make test` coverage.
