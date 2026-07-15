# Feature Specification: singctl CLI/TUI for SingularityApp

**Feature Branch**: `001-singctl-cli-tui`

**Created**: 2026-07-15

**Status**: Draft

**Input**: Исходное ТЗ `docs/tz/singularityapp-cli-tui-tz.md` + ограничения constitution (Go, OpenAPI codegen) + актуальный OpenAPI SingularityApp v2.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Настройка доступа и проверка API (Priority: P1)

Технический пользователь Pro/Elite сохраняет Bearer-токен, проверяет конфигурацию и убеждается, что клиент может достучаться до API.

**Why this priority**: Без рабочей авторизации остальные сценарии невозможны.

**Independent Test**: Установить токен, выполнить `config validate` / эквивалент list с лимитом 1; при неверном токене получить понятную ошибку 401.

**Acceptance Scenarios**:

1. **Given** отсутствует конфиг, **When** пользователь задаёт токен через CLI, **Then** токен сохраняется в стандартном расположении конфига и маскируется при просмотре.
2. **Given** валидный токен, **When** выполняется проверка подключения, **Then** команда завершается успешно (exit 0).
3. **Given** невалидный/отозванный токен, **When** выполняется любая API-команда, **Then** выводится сообщение с подсказкой обновить токен (exit ≠ 0).

---

### User Story 2 - CRUD задач в CLI (Priority: P1)

Пользователь управляет задачами из терминала: список с фильтрами, получение, создание, обновление, удаление, архив/корзина, чек-листы, перемещение в канбан.

**Why this priority**: Задачи — центральная сущность продукта и MVP по исходному ТЗ.

**Independent Test**: Полный цикл list → create → get → update → archive/trash → checklist → delete на тестовом аккаунте или моке.

**Acceptance Scenarios**:

1. **Given** валидный токен, **When** `task list` с фильтрами проекта/дат/пагинации, **Then** выводится таблица (или JSON) задач.
2. **Given** обязательный `--title`, **When** `task create`, **Then** создаётся задача и возвращается её ID.
3. **Given** существующая задача, **When** archive/trash/update/delete, **Then** состояние на сервере меняется согласно API.
4. **Given** задача, **When** операции checklist / move в колонку, **Then** вызываются соответствующие endpoints чек-листа и kanban-task-status.

---

### User Story 3 - Форматы вывода и скриптуемость (Priority: P1)

Пользователь использует CLI в пайпах и автоматизации.

**Why this priority**: Целевая аудитория — терминальный workflow; без pipe-режима ценность CLI низкая.

**Independent Test**: `task list --output json | jq` и проверка exit codes / отключения цвета в non-TTY.

**Acceptance Scenarios**:

1. **Given** команда list, **When** `--output json|yaml|csv|table`, **Then** stdout соответствует формату.
2. **Given** non-TTY stdout, **When** команда без явного цвета, **Then** ANSI-цвета отключены.
3. **Given** ошибки API / конфига / not found, **When** команда завершается, **Then** exit codes: 0 успех, 1 API, 2 конфиг, 3 not found.

---

### User Story 4 - Проекты, привычки, теги, время в CLI (Priority: P2)

Пользователь управляет остальными сущностями API теми же паттернами команд.

**Why this priority**: Расширяет покрытие API после MVP задач.

**Independent Test**: Для каждой сущности — list/get/create/update/delete (+ специфичные операции: habit progress, time bulk delete, project columns/sections).

**Acceptance Scenarios**:

1. **Given** валидный токен, **When** CRUD `project` / `habit` / `tag` / `time`, **Then** операции соответствуют OpenAPI.
2. **Given** проект, **When** управление колонками и секциями, **Then** используются kanban-status и task-group endpoints.
3. **Given** привычка и дата, **When** отметка прогресса, **Then** создаётся/обновляется запись habit-progress.
4. **Given** записи времени, **When** bulk delete, **Then** удаляются выбранные записи через API bulk-операцию.

---

### User Story 5 - Интерактивный TUI (Priority: P2–P3)

Пользователь запускает `singctl tui` (или `singctl` без аргументов) и управляет сущностями клавиатурой.

**Why this priority**: Дифференцирует продукт; базовый TUI задач/проектов — Phase 2, полный — Phase 3 по ТЗ.

**Independent Test**: Запуск TUI с мок-клиентом: навигация панелей, создание/редактирование задачи, refresh.

**Acceptance Scenarios**:

1. **Given** валидный токен, **When** запуск TUI, **Then** доступны разделы Задачи / Проекты / Привычки / Теги / Время.
2. **Given** раздел Задачи, **When** хоткеи create/edit/archive/checklist, **Then** изменения уходят через тот же API-слой, что и CLI.
3. **Given** API ошибка 401/5xx, **When** пользователь в TUI, **Then** показывается баннер/статус без падения процесса.
4. **Given** `vi_keys: true`, **When** навигация по спискам, **Then** работают j/k/g/G и page scroll.

---

### User Story 6 - DX: completions, алиасы, quick-add, кэш (Priority: P3)

Пользователь ускоряет работу через shell completion, короткие алиасы, `quick-add` синтаксис и локальный кэш справочников.

**Why this priority**: Удобство после стабильного CRUD/TUI.

**Independent Test**: Генерация completion-скрипта; разбор строки quick-add; TTL-кэш проектов/тегов.

**Acceptance Scenarios**:

1. **Given** shell bash/zsh/fish, **When** `completion <shell>`, **Then** выводится валидный скрипт автодополнения.
2. **Given** алиасы `t|p|h|ti`, **When** вызов, **Then** эквивалент полным командам.
3. **Given** строка `quick-add` с `@project #tag !priority date`, **When** парсинг, **Then** создаётся задача с соответствующими полями (при резолве имён через кэш/API).
4. **Given** кэш проектов/тегов, **When** мутация, **Then** кэш инвалидируется.

---

### Edge Cases

- Отсутствует токен → wizard / явная ошибка конфига (exit 2), без паники.
- 429 Too Many Requests → до 3 retry с exponential backoff; затем ошибка.
- 422 → показать тело/сообщение валидации; в TUI — inline в форме.
- Неверный формат даты → подсказка ожидаемого формата.
- Попытка создать recurring task через CLI → предупреждение: API не поддерживает создание recurrence.
- Pipe + таблица: заголовки стабильны; JSON — машиночитаемый массив/объект.
- Удалённый/архивный include-флаги не ломают default list.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Система MUST предоставлять CLI-бинарник `singctl` с ресурсами `config`, `task`, `project`, `habit`, `tag`, `time`, `tui`, `completion`, `quick-add`.
- **FR-002**: Система MUST читать конфигурацию с приоритетом: `--config` → `$XDG_CONFIG_HOME/singctl/config.yaml` → `~/.config/singctl/config.yaml` → `./.singctl.yaml`.
- **FR-003**: Система MUST поддерживать глобальные флаги `--token`, `--output/-o`, `--no-color`, `--debug`.
- **FR-004**: Система MUST аутентифицироваться к API через Bearer-токен в заголовке `Authorization`.
- **FR-005**: Система MUST выполнять CRUD задач, проектов, привычек, тегов, time-stat, checklist-item, kanban-status, kanban-task-status, task-group, habit-progress согласно OpenAPI.
- **FR-006**: Система MUST поддерживать форматы вывода table, json, yaml, csv.
- **FR-007**: Система MUST использовать единый API-клиент для CLI и TUI.
- **FR-008**: Система MUST генерировать API-модели/клиент из OpenAPI, а не поддерживать ручные DTO сущностей.
- **FR-009**: Система MUST обрабатывать HTTP 401/403/404/422/429/5xx с предсказуемым UX (CLI сообщения / TUI баннеры) и exit codes.
- **FR-010**: TUI MUST поддерживать навигацию Tab/цифры 1–5, `?`, `q`, `/`, `r`, `Esc` и разделы из ТЗ (задачи — приоритет Phase 2).
- **FR-011**: Система MUST отключать цвет при non-TTY, если не переопределено явно.
- **FR-012**: Система MUST NOT реализовывать webhooks, offline sync, управление чужими shared-проектами или редактирование токена на сервере.

### Key Entities *(include if feature involves data)*

- **Task** — задача пользователя (title, priority, dates, project, checklist, kanban link).
- **Project** — проект/блокнот; связанные секции (task-group) и kanban-status.
- **Habit** + **HabitDailyProgress** — привычка и дневная отметка прогресса.
- **Tag** — иерархический тег.
- **TimeStat** — запись затраченного времени, опционально связанная с задачей.
- **ChecklistItem** — пункт чек-листа задачи.
- **Config** — локальные настройки клиента (base URL, token, output, tui).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Пользователь с валидным токеном выполняет list+create+update задачи менее чем за 3 команды CLI без открытия GUI SingularityApp.
- **SC-002**: Все endpoints из `docs/api/openapi.yaml`, нужные для сущностей ТЗ, доступны через CLI-команды или явны как out-of-scope в tasks.
- **SC-003**: Pipe `singctl task list -o json | jq 'length'` работает на типовом аккаунте без ручной очистки ANSI.
- **SC-004**: TUI базового раздела задач запускается одной командой и позволяет создать задачу end-to-end.
- **SC-005**: Повторная генерация клиента из обновлённого OpenAPI описана в `docs/` и выполняется одной документированной командой.

## Assumptions

- Целевые пользователи имеют тариф Pro/Elite и умеют создать токен в личном кабинете.
- Base URL по умолчанию: `https://api.singularity-app.com` (уточнено относительно черновика `api.singularity-app.ru` в ТЗ).
- Имя бинарника: `singctl` (рабочее имя из ТЗ); репозиторий может называться `sa-cli`.
- Расхождения ТЗ ↔ OpenAPI (например bulk delete time-stat методом DELETE, habit-progress отдельным ресурсом) разрешаются в пользу OpenAPI; см. `research.md`.
- Дистрибуция goreleaser/Homebrew/deb — Phase 3, не блокирует MVP.
