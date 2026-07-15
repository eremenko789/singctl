# Research: singctl CLI/TUI

**Feature**: `001-singctl-cli-tui` | **Date**: 2026-07-15

## 1. Language and UI stack

**Decision**: Go + Cobra/Viper + Charmbracelet (bubbletea/lipgloss/bubbles).

**Rationale**: Совпадает с рекомендованным стеком ТЗ; один бинарник; зрелая TUI-экосистема; хорошая поддержка codegen HTTP-клиентов.

**Alternatives considered**: Python (click+textual) — быстрее прототип, хуже дистрибуция single binary; отклонён constitution.

## 2. API client generation

**Decision**: Снимок OpenAPI из `https://api.singularity-app.com/v2/api-json` → `oapi-codegen` → `internal/apiclient`.

**Rationale**: Требование пользователя и constitution; снижает drift моделей; Swagger UI уже публикует OpenAPI 3.0.

**Alternatives considered**: ручные DTO (как в черновом дереве ТЗ `internal/api/*.go`) — отклонено; `openapi-generator` go — допустимый запасной вариант, см. `docs/openapi-codegen.md`.

## 3. API base URL

**Decision**: Default `https://api.singularity-app.com` (не `api.singularity-app.ru` из черновика ТЗ).

**Rationale**: Подтверждено wiki и рабочим Swagger/OpenAPI host.

## 4. Расхождения ТЗ ↔ актуальный OpenAPI

Зафиксировать в реализации поведение по OpenAPI:

| Тема в ТЗ | Факт OpenAPI / решение |
|---|---|
| `POST /v2/time-stat/delete-bulk` | Bulk delete: `DELETE /v2/time-stat` (`TimeStatController_deleteBulk`) |
| `habit track` как действие на habit | Отдельный ресурс `/v2/habit-progress` (CRUD) |
| time `--type TIME/BREAK` | В схемах — поля `source` / `secondsPassed` / `relatedTaskId`; CLI-флаги маппить на реальные поля после чтения схемы |
| `task move` через POST kanban-task-status | Да: create/update `KanbanTaskStatus` (link task↔status) |
| Секции проекта | `/v2/task-group` |
| Канбан-колонки | `/v2/kanban-status` |
| Токен в конфиге как `Bearer xxx` | Хранить raw token; префикс `Bearer` добавляет клиент |

**Rationale**: Машиночитаемый контракт надёжнее wiki-скриншотов и чернового ТЗ.

## 5. Config locations

**Decision**: Как в ТЗ (flag → XDG → `~/.config/singctl` → `./.singctl.yaml`).

**Rationale**: Стандарт для CLI на Linux/macOS; Windows — через аналогичные пути viper/xdg.

## 6. Phasing

**Decision**: Сохранить фазы ТЗ:

1. MVP: codegen client, config, task CRUD, table/json
2. Остальные CLI-ресурсы + базовый TUI задач/проектов + completions
3. Полный TUI, quick-add, cache, goreleaser

**Rationale**: Независимо тестируемые user stories P1→P3.

## 7. Testing strategy

**Decision**: httptest/мок на адаптерах; золотые файлы вывода table/json опционально; integration с реальным API только opt-in через env token.

**Rationale**: Публичный API требует пользовательский токен; CI без секретов должен быть зелёным.

## 8. Open questions (не блокируют plan)

- Точные enum значения `source` у TimeStat — уточнить из схемы/примеров при реализации time-команд.
- Нужен ли отдельный `man` page в MVP — отложить до Phase 3 (документация ТЗ §13.2).
