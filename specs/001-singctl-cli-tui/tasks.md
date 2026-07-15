# Tasks: singctl CLI/TUI

**Input**: Design documents from `/specs/001-singctl-cli-tui/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Включены для API-адаптеров и критичных парсеров (требование качества из ТЗ §13.1).

**Organization**: По user stories; Phase 1–2 блокируют stories.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: можно параллелить
- **[Story]**: US1…US6

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Каркас Go-модуля

- [ ] T001 Инициализировать Go-модуль в корне репозитория (`go.mod`), `cmd/singctl/main.go`
- [ ] T002 [P] Добавить зависимости cobra/viper и заготовки каталогов `internal/{config,api,apiclient,output,tui}`
- [ ] T003 [P] Добавить `api/oapi-codegen.yaml` и `//go:generate` по `docs/openapi-codegen.md`
- [ ] T004 [P] Добавить `config.yaml.example` согласно spec FR-002/конфигу ТЗ
- [ ] T005 Обновить корневой `README.md`: установка, конфиг, ссылки на `docs/` и `specs/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Codegen-клиент, конфиг, ошибки, вывод — до любых user stories

- [ ] T006 Обновить/подтвердить снимок `docs/api/openapi.yaml` и скопировать в `specs/001-singctl-cli-tui/contracts/`
- [ ] T007 Сгенерировать `internal/apiclient/*.gen.go` через oapi-codegen
- [ ] T008 Реализовать `internal/config` (пути, load/save, mask token, `--config` override)
- [ ] T009 Реализовать `internal/api` обёртку: base URL, Bearer auth, timeout, debug logging без утечки токена
- [ ] T010 [P] Добавить retry для HTTP 429 (3 попытки, exponential backoff) в `internal/api`
- [ ] T011 [P] Маппинг ошибок API → типы + exit codes (1/2/3) в `internal/api` / CLI root
- [ ] T012 [P] Реализовать `internal/output` для table/json/yaml/csv + auto `--no-color` на non-TTY
- [ ] T013 Корневая cobra-команда `singctl` с глобальными флагами FR-003
- [ ] T014 Unit-тесты на config paths и error/exit mapping (`internal/config`, `internal/api`)

**Checkpoint**: Foundation ready

---

## Phase 3: User Story 1 - Настройка доступа (Priority: P1) 🎯 MVP

**Goal**: set-token / show / validate / set

**Independent Test**: quickstart §A

- [ ] T015 [US1] Команды `singctl config set-token|show|validate|set` в cobra
- [ ] T016 [US1] `validate` делает лёгкий GET (например `/v2/project?maxCount=1`)
- [ ] T017 [US1] Тесты команды config с мок HTTP

**Checkpoint**: US1 done

---

## Phase 4: User Story 2 - CRUD задач (Priority: P1) 🎯 MVP

**Goal**: полный CLI task + checklist + move

**Independent Test**: quickstart §B

- [ ] T018 [P] [US2] Фасады task в `internal/api` над generated client
- [ ] T019 [US2] `singctl task list|get|create|update|delete` с флагами из ТЗ/OpenAPI
- [ ] T020 [US2] `singctl task archive|trash` через PATCH дат
- [ ] T021 [P] [US2] `singctl task checklist *` → `/v2/checklist-item`
- [ ] T022 [US2] `singctl task move` → `/v2/kanban-task-status`
- [ ] T023 [US2] Unit-тесты list filters и archive/trash mapping на httptest

**Checkpoint**: MVP CLI задач готов

---

## Phase 5: User Story 3 - Форматы и pipe (Priority: P1) 🎯 MVP

**Goal**: стабильный scripting UX

**Independent Test**: json pipe + exit codes

- [ ] T024 [US3] Проверить все task-команды на `-o json|yaml|csv|table`
- [ ] T025 [US3] Интеграционные smoke-тесты output + non-TTY color off
- [ ] T026 [US3] Документировать примеры pipe в README

**Checkpoint**: MVP (US1–US3) можно релизить как pre-1.0 CLI

---

## Phase 6: User Story 4 - Остальные сущности CLI (Priority: P2)

**Goal**: project/habit/tag/time (+ columns/sections/progress/bulk)

**Independent Test**: quickstart §C

- [ ] T027 [P] [US4] `singctl project` CRUD + `column` + `section`
- [ ] T028 [P] [US4] `singctl habit` CRUD + `track` через habit-progress
- [ ] T029 [P] [US4] `singctl tag` CRUD (hierarchy flags)
- [ ] T030 [US4] `singctl time` list/add/update/delete/delete-bulk по OpenAPI
- [ ] T031 [US4] Unit-тесты фасадов US4 с моками

**Checkpoint**: CLI покрывает сущности ТЗ

---

## Phase 7: User Story 5 - TUI (Priority: P2–P3)

**Goal**: bubbletea app; сначала задачи/проекты, затем остальные разделы

**Independent Test**: quickstart §D

- [ ] T032 [US5] Каркас `internal/tui` + `singctl tui` / default без args
- [ ] T033 [US5] Раздел Задачи (список, детали, create/edit/archive, refresh)
- [ ] T034 [US5] Раздел Проекты (+ базовый kanban view)
- [ ] T035 [P] [US5] Разделы Привычки / Теги / Время
- [ ] T036 [US5] Глобальные хоткеи + vi_keys + error banners
- [ ] T037 [US5] Model-тесты ключевых tea.Msg сценариев

**Checkpoint**: TUI usable

---

## Phase 8: User Story 6 - DX enhancements (Priority: P3)

**Goal**: completions, aliases, quick-add, cache

- [ ] T038 [P] [US6] `singctl completion bash|zsh|fish`
- [ ] T039 [P] [US6] Алиасы `t|p|h|ti`
- [ ] T040 [US6] `internal/quickadd` parser + `singctl quick-add`
- [ ] T041 [US6] `internal/cache` TTL 5m для projects/tags + инвалидация на мутациях
- [ ] T042 [US6] Интерактивный prompt при create без обязательных флагов (fzf-like / survey)

---

## Phase 9: Polish & Distribution

- [ ] T043 [P] man page `singctl.1` и расширенный `--help`
- [ ] T044 [P] `CHANGELOG.md`
- [ ] T045 Goreleaser конфиг (linux/mac/windows amd64+arm64)
- [ ] T046 [P] Опциональный Dockerfile
- [ ] T047 Прогон `/speckit.analyze`-эквивалента: сверить tasks ↔ FR ↔ OpenAPI paths

---

## Dependencies (stories)

- Phase 1–2 → все US
- US1 → US2/US4/US5 (нужен config/auth)
- US2 → US3 (можно частично параллельно после T012)
- US4 после Phase 2 (параллельно с поздним US3)
- US5 после фасадов нужных сущностей (минимум task+project)
- US6 после стабильного CLI

## Parallel examples

- T027/T028/T029 после T009
- T038/T039 после корневой cobra

## MVP scope (recommended first delivery)

Только: T001–T026 (Setup + Foundation + US1 + US2 + US3).
