# ТЗ: CLI/TUI-клиент для SingularityApp REST API

## 1. Общие сведения

**Назначение продукта:** Консольный клиент (CLI + TUI) для управления задачами, проектами, привычками, тегами и записями времени через [REST API SingularityApp](https://singularity-app.ru/wiki/api/).[^1]

**Целевая аудитория:** Технические пользователи SingularityApp (тарифы Pro/Elite), предпочитающие терминальный workflow.[^1][^2]

**Рабочее название:** `singctl` (или `sa-cli`)

***

## 2. Контекст API

SingularityApp предоставляет REST API с базовым префиксом `/v2/`.[^1] Авторизация — Bearer-токен в заголовке `Authorization`.[^1] Поддерживаются операции CRUD (GET, POST, PATCH, DELETE). Вебхуки, подписки на события и потоки не поддерживаются.[^1]

**Сущности API:**

| Сущность | Endpoint | Описание |
|---|---|---|
| Задачи | `/v2/task` | CRUD + archive/trash через PATCH |
| Чек-листы | `/v2/checklist-item` | полный CRUD (parent = task) |
| Проекты | `/v2/project` | CRUD |
| Секции | `/v2/task-group` | полный CRUD (группы задач в проекте) |
| Канбан-колонки | `/v2/kanban-status` | полный CRUD |
| Канбан-связь задачи | `/v2/kanban-task-status` | полный CRUD (move = create/update) |
| Привычки | `/v2/habit` | CRUD |
| Прогресс привычек | `/v2/habit-progress` | полный CRUD (track = create/update) |
| Теги | `/v2/tag` | CRUD, иерархия |
| Записи времени | `/v2/time-stat` | CRUD + bulk DELETE по фильтрам |

Каноническая матрица всех operations: `docs/api/coverage.md` (сверка с OpenAPI).

***

## 3. Стек технологий

**Рекомендуемый язык:** Go — нативная компиляция в один бинарник, отличная поддержка TUI (`bubbletea`/`lipgloss`) и HTTP-клиентов.

**Альтернатива:** Python (`click` + `textual`) — если важна скорость разработки прототипа.

**Ключевые библиотеки (Go):**

| Назначение | Библиотека |
|---|---|
| TUI-фреймворк | `github.com/charmbracelet/bubbletea` |
| Стилизация | `github.com/charmbracelet/lipgloss` |
| Таблицы/списки | `github.com/charmbracelet/bubbles` |
| CLI-парсинг | `github.com/spf13/cobra` |
| HTTP-клиент | `net/http` (stdlib) |
| Конфигурация | `github.com/spf13/viper` |
| Вывод таблиц CLI | `github.com/olekukonko/tablewriter` |

***

## 4. Архитектура

```
singctl/
├── cmd/                   # CLI команды (cobra)
│   ├── root.go
│   ├── task/
│   ├── project/
│   ├── habit/
│   ├── tag/
│   ├── time/
│   └── tui.go             # Запуск TUI-режима
├── internal/
│   ├── apiclient/         # generated OpenAPI client (не вручную)
│   ├── api/               # тонкие адаптеры над codegen
│   ├── config/            # Конфигурация (токен, base URL)
│   └── tui/               # Bubbletea-компоненты
├── api/
│   └── oapi-codegen.yaml  # конфиг генератора
├── Makefile               # регулярные задачи (+ .env)
├── config.yaml.example
└── main.go / cmd/singctl/
```

**Принцип работы:** CLI-режим вызывается напрямую (`singctl task list`), TUI-режим — через `singctl tui` или без аргументов. Оба режима используют один и тот же API-слой над OpenAPI-клиентом. Покрытие всех REST operations — `docs/api/coverage.md`.

***

## 5. Конфигурация

### 5.1 Файл конфигурации

Расположение (приоритет убывает):
1. Флаг `--config /path/to/config.yaml`
2. `./.singctl.yaml`
3. `$XDG_CONFIG_HOME/singctl/config.yaml`
4. `~/.config/singctl/config.yaml`

**Формат `config.yaml`:**
```yaml
api:
  base_url: "https://api.singularity-app.ru"   # уточнить в Swagger
  token: "Bearer xxxx..."
  timeout: 30s

output:
  format: table   # table | json | csv | yaml
  color: true
  date_format: "2006-01-02"

tui:
  theme: dark   # dark | light
  vi_keys: true
  refresh_interval: 0   # 0 = no auto-refresh (API не поддерживает подписки)
```

### 5.2 Управление токеном через CLI

```bash
singctl config set-token <TOKEN>       # сохранить токен
singctl config show                    # показать текущую конфигурацию (токен маскируется)
singctl config validate                # проверить подключение к API
```

***

## 6. CLI-спецификация

Общий формат: `singctl <ресурс> <действие> [флаги]`

Глобальные флаги:

| Флаг | Описание |
|---|---|
| `--config` | Путь к файлу конфигурации |
| `--token` | Токен (override конфига) |
| `--output, -o` | Формат: `table`, `json`, `yaml`, `csv` |
| `--no-color` | Отключить цвет |
| `--debug` | Verbose HTTP-лог |

***

### 6.1 Задачи (`task`)

```
singctl task list [флаги]
```

| Флаг | Тип | API-поле | Описание |
|---|---|---|---|
| `--project` | string | `projectId` | Фильтр по проекту |
| `--parent` | string | `parent` | Подзадачи указанной задачи |
| `--from` | date | `startDateFrom` | Дата начала (ISO) |
| `--to` | date | `startDateTo` | Дата окончания (ISO) |
| `--archived` | bool | `includeArchived` | Включать архивные |
| `--removed` | bool | `includeRemoved` | Включать удалённые |
| `--limit` | int | `maxCount` | Макс. количество |
| `--offset` | int | `offset` | Смещение (пагинация) |
| `--all-recurrence` | bool | `includeAllRecurrenceInstances` | Все экземпляры повторяющихся |

```
singctl task get <ID>
singctl task create --title "Название" [флаги]
```

| Флаг | Тип | API-поле | Обязательный |
|---|---|---|---|
| `--title` | string | `title` | ✅ |
| `--start` | date | `start` | ❌ |
| `--note` | string | `note` | ❌ |
| `--priority` | 0/1/2 | `priority` | ❌ |
| `--is-note` | bool | `isNote` | ❌ |
| `--archive-date` | date | `journalDate` | ❌ |
| `--delete-date` | date | `deleteDate` | ❌ |

```
singctl task update <ID> [флаги]        # те же флаги что create, все опциональны
singctl task delete <ID>                # безвозвратное удаление
singctl task archive <ID> [--date DATE] # PATCH journalDate
singctl task trash <ID> [--date DATE]   # PATCH deleteDate
singctl task move <ID> --column <COLUMN_ID>  # POST /v2/kanban-task-status
```

**Чек-листы** (`/v2/checklist-item`):
```
singctl task checklist list <TASK_ID>
singctl task checklist get <CHECKLIST_ITEM_ID>
singctl task checklist add <TASK_ID> --title "Пункт"
singctl task checklist update <CHECKLIST_ITEM_ID> [--title ...] [--done]
singctl task checklist delete <CHECKLIST_ITEM_ID>
```

**Канбан-связь задачи** (`/v2/kanban-task-status`, полный CRUD; `move` — UX над create/update):
```
singctl task kanban list [--task <ID>] [--status <COLUMN_ID>]
singctl task kanban get <LINK_ID>
singctl task kanban create --task <ID> --column <COLUMN_ID> [--order N]
singctl task kanban update <LINK_ID> [--task ...] [--column ...] [--order ...]
singctl task kanban delete <LINK_ID>
singctl task move <ID> --column <COLUMN_ID>   # create/update link
```

***

### 6.2 Проекты (`project`)

```
singctl project list [--archived] [--removed] [--limit N] [--offset N]
singctl project get <ID>
singctl project create --title "Название" [флаги]
```

| Флаг | Тип | API-поле |
|---|---|---|
| `--title` | string | `title` |
| `--note` | string | `note` |
| `--notebook` | bool | `isNotebook` |
| `--emoji` | string | `emoji` (hex, напр. `1f49e`) |
| `--color` | string | `color` (HEX) |

```
singctl project update <ID> [флаги]
singctl project delete <ID>
```

**Канбан-колонки** (`/v2/kanban-status`):
```
singctl project column list <PROJECT_ID>
singctl project column get <COLUMN_ID>
singctl project column create <PROJECT_ID> --title "Название"
singctl project column update <COLUMN_ID> --title "Новое название"
singctl project column delete <COLUMN_ID>
```

**Секции** (`/v2/task-group`):
```
singctl project section list <PROJECT_ID>
singctl project section get <SECTION_ID>
singctl project section create <PROJECT_ID> --title "Название"
singctl project section update <SECTION_ID> --title "Новое название"
singctl project section delete <SECTION_ID>
```

***

### 6.3 Привычки (`habit`)

```
singctl habit list [--limit N]
singctl habit get <ID>
singctl habit create --title "Пить воду" [--description "..."] [--color teal]
singctl habit update <ID> [--title ...] [--description ...] [--color ...]
singctl habit delete <ID>
singctl habit track <ID> --date 2025-11-28 --progress <0|1|2>
```

Значения `--progress`:[^1]
- `0` — стандартное, без изменения прогресса
- `1` — НЕ выполнено (серия сохраняется)
- `2` — выполнено

**Прогресс привычек** (ресурс OpenAPI `/v2/habit-progress`, полный CRUD):

```
singctl habit progress list [--habit <ID>] [--from DATE] [--to DATE] [--limit N] [--offset N]
singctl habit progress get <ID>
singctl habit progress create --habit <ID> --date DATE --progress <0|1|2>
singctl habit progress update <ID> [--habit ...] [--date ...] [--progress ...]
singctl habit progress delete <ID>
```

`habit track` — UX-алиас над `habit progress create` (и при необходимости update существующей записи за дату).

***

### 6.4 Теги (`tag`)

```
singctl tag list [--parent <ID>] [--removed] [--limit N] [--offset N]
singctl tag get <ID>
singctl tag create --title "Работа" [--parent <ID>] [--hotkey 57]
singctl tag update <ID> [--title ...] [--parent ...] [--hotkey ...]
singctl tag delete <ID>
```

***

### 6.5 Время (`time`)

```
singctl time list [флаги]
```

| Флаг | Тип | API-поле |
|---|---|---|
| `--task` | string | `relatedTaskId` |
| `--from` | datetime | `dateFrom` |
| `--to` | datetime | `dateTo` |
| `--limit` | int | `maxCount` |
| `--offset` | int | `offset` |

```
singctl time get <ID>
singctl time add --start "2025-11-28T09:00:00Z" --duration 3600 [--task <ID>] [--source ...]
singctl time update <ID> [--start ...] [--duration ...] [--task ...] [--source ...]
singctl time delete <ID>
singctl time delete-bulk [--from ...] [--to ...] [--task ...]   # DELETE /v2/time-stat (фильтры OpenAPI)
```

Поля и фильтры — по актуальному OpenAPI (`secondsPassed`, `relatedTaskId`, `source`); черновик `--type TIME|BREAK` уточняется при plan/codegen.

Полная матрица всех 51 operations: [`docs/api/coverage.md`](../api/coverage.md).

***

### 6.6 Конфигурация (`config`)

```
singctl config set-token <TOKEN>
singctl config show
singctl config validate
singctl config set <key> <value>   # Установить произвольный ключ
```

***

## 7. TUI-спецификация

Запуск: `singctl tui` или просто `singctl` без аргументов.

### 7.1 Навигация

**Глобальные хоткеи:**

| Клавиша | Действие |
|---|---|
| `Tab` / `Shift+Tab` | Переключение между панелями |
| `1`–`5` | Быстрый переход к разделу (Задачи / Проекты / Привычки / Теги / Время) |
| `?` | Показать справку по хоткеям |
| `q` / `Ctrl+C` | Выход |
| `/` | Глобальный поиск/фильтр |
| `r` | Обновить текущий вид (повторный GET) |
| `Esc` | Закрыть модальное окно / отмена |

**Vi-режим навигации** (если `vi_keys: true` в конфиге):

| Клавиша | Действие |
|---|---|
| `j` / `k` | Вниз / вверх по списку |
| `g` / `G` | В начало / конец списка |
| `Ctrl+d` / `Ctrl+u` | Пролистать вниз / вверх |

***

### 7.2 Раздел «Задачи»

**Компоненты:**
- Левая панель: дерево проектов (с возможностью выбора фильтра)
- Центральная панель: список задач с колонками: ID, Заголовок, Приоритет, Дата начала, Теги
- Правая панель: детали выбранной задачи (title, note, checklist)

**Хоткеи раздела:**

| Клавиша | Действие |
|---|---|
| `n` | Создать новую задачу (форма) |
| `e` | Редактировать выбранную задачу |
| `d` | Удалить (с подтверждением) |
| `a` | Архивировать |
| `Space` | Отметить как завершённую (PATCH `journalDate` = сегодня) |
| `m` | Переместить в канбан-колонку (диалог выбора) |
| `c` | Открыть чек-лист задачи |
| `t` | Добавить запись времени к задаче |
| `Enter` | Открыть детали задачи |
| `f` | Открыть панель фильтров |

**Inline-форма создания/редактирования:**
- title (text input)
- start date (date picker)
- priority (radio: Высокий / Обычный / Низкий)
- note (textarea)
- is_note (toggle)

***

### 7.3 Раздел «Проекты»

**Компоненты:**
- Список проектов с эмодзи, цветом, именем
- При выборе проекта — канбан-вид задач по колонкам (горизонтальный скролл)
- Панель управления секциями

**Хоткеи:**

| Клавиша | Действие |
|---|---|
| `n` | Новый проект |
| `e` | Редактировать проект |
| `d` | Удалить проект |
| `+` | Добавить канбан-колонку |
| `s` | Добавить секцию |
| `Enter` | Открыть проект (канбан/список задач) |

***

### 7.4 Раздел «Привычки»

**Компоненты:**
- Список привычек с цветовыми метками
- Мини-трекер: таблица дней текущей недели/месяца (X — выполнено, O — пропущено, · — без данных)
- Детали привычки: описание, текущая серия

**Хоткеи:**

| Клавиша | Действие |
|---|---|
| `n` | Новая привычка |
| `e` | Редактировать |
| `d` | Удалить |
| `Enter` / `Space` | Отметить привычку сегодня (progress=2) |
| `x` | Отметить как НЕ выполненную (progress=1) |
| `0` | Сбросить отметку (progress=0) |
| `←` / `→` | Выбрать дату для отметки |

***

### 7.5 Раздел «Теги»

**Компоненты:**
- Дерево тегов (с иерархией parent → children)
- Inline-редактирование (горячая клавиша, название)

**Хоткеи:**

| Клавиша | Действие |
|---|---|
| `n` | Новый тег |
| `N` | Новый дочерний тег (под выбранным) |
| `e` | Редактировать |
| `d` | Удалить |

***

### 7.6 Раздел «Время»

**Компоненты:**
- Список записей времени с фильтром по задаче и датам
- Суммарная статистика (общее время за день / неделю — вычисляется клиентом)
- Форма добавления записи

**Хоткеи:**

| Клавиша | Действие |
|---|---|
| `n` | Новая запись времени |
| `e` | Редактировать запись |
| `d` | Удалить запись |
| `D` | Массовое удаление (multi-select, затем `D`) |
| `f` | Фильтр по задаче / дате / типу |

***

## 8. Обработка ошибок

### 8.1 HTTP-ошибки

| HTTP-код | Поведение CLI | Поведение TUI |
|---|---|---|
| 401 | `Error: invalid token. Run 'singctl config set-token'` | Красный баннер + подсказка |
| 403 | `Error: insufficient token permissions` | Красный баннер |
| 404 | `Error: entity not found: <ID>` | Предупреждение в статусбаре |
| 422 | Вывод сообщения из тела ответа | Ошибка в форме (inline) |
| 429 | Авто-retry с exponential backoff (3 попытки) | Индикатор ожидания |
| 5xx | `Error: server error, retry later` | Красный баннер |

### 8.2 Клиентские ошибки

- Отсутствующий токен → пошаговый wizard настройки (`singctl config set-token`)
- Неверный формат даты → подсказка `Expected: YYYY-MM-DD`
- Попытка создать повторяющуюся задачу → предупреждение (API не поддерживает)[^1]

***

## 9. Форматы вывода CLI

### 9.1 Таблица (default)

```
ID          TITLE               PRIORITY  START       PROJECT
T-abc123    Купить продукты     high      2025-11-28  Личное
T-def456    Созвон с командой   normal    2025-11-29  Работа
```

### 9.2 JSON (`--output json`)

Сырой ответ API или массив объектов в зависимости от команды.

### 9.3 YAML (`--output yaml`)

Удобно для pipe в другие инструменты.

### 9.4 CSV (`--output csv`)

Для экспорта в spreadsheets.

***

## 10. Pipe-режим и скриптуемость

Все команды должны поддерживать pipe. Примеры использования:

```bash
# Все задачи без дедлайна → в JSON-файл
singctl task list --output json > tasks.json

# Создать задачи из файла построчно
cat tasks.txt | xargs -I{} singctl task create --title "{}"

# Архивировать все задачи проекта
singctl task list --project P-123 --output json \
  | jq -r '.[].id' \
  | xargs -I{} singctl task archive {}

# Статистика по времени за неделю
singctl time list --from 2025-11-25 --to 2025-12-01 --output csv \
  | awk -F, 'NR>1 {sum+=$4} END {print sum/3600 " hours"}'
```

**Требования к pipe-режиму:**
- При наличии stdin/non-tty: автоматически отключить цвет (`--no-color`)
- Exit codes: `0` — успех, `1` — ошибка API, `2` — ошибка конфигурации, `3` — not found

***

## 11. Дополнительные функции

### 11.1 Автодополнение shell

```bash
singctl completion bash   >> ~/.bashrc
singctl completion zsh    >> ~/.zshrc
singctl completion fish   >> ~/.config/fish/completions/singctl.fish
```

### 11.2 Алиасы команд

Сокращения для частых операций:

| Алиас | Полная команда |
|---|---|
| `singctl t` | `singctl task` |
| `singctl p` | `singctl project` |
| `singctl h` | `singctl habit` |
| `singctl ti` | `singctl time` |

### 11.3 Интерактивный режим форм (fzf-like)

При вызове `singctl task create` без `--title` — интерактивный prompt с автодополнением названий проектов и тегов.

### 11.4 Локальный кэш

- Кэш списков проектов и тегов для автодополнения (TTL 5 минут, хранится в `~/.cache/singctl/`)
- Инвалидация: при любой мутирующей операции (POST/PATCH/DELETE)

### 11.5 Макрос `quick-add`

```bash
singctl quick-add "Купить молоко @Личное #продукты !high 2025-11-28"
```

Парсинг синтаксиса:
- `@ProjectName` → `projectId`
- `#TagName` → тег
- `!high|!normal|!low` → приоритет
- `YYYY-MM-DD` → `start`

***

## 12. Что API НЕ поддерживает (ограничения)

Следующие функции реализовать невозможно из-за ограничений API[^1]:

| Функция | Причина |
|---|---|
| Real-time обновления / уведомления | API не поддерживает webhooks/SSE |
| Создание повторяющихся задач | Ограничение API |
| Управление совместными проектами | API возвращает только личные проекты |
| Редактирование токена | Только создание/удаление через личный кабинет |
| Офлайн-режим с синхронизацией | Нет конфликт-резолюции на стороне API |

***

## 13. Требования к качеству

### 13.1 Тестирование

- Unit-тесты: `internal/api/*` — мок HTTP-сервер (100% покрытие полей)
- Integration-тесты: `cmd/*` — real API с тестовым токеном (CI/CD)
- TUI-тесты: bubbletea `tea.Msg`-based тесты моделей

### 13.2 Документация

- `--help` для каждой команды и субкоманды
- `man`-страница (`singctl.1`)
- `README.md` с примерами
- `CHANGELOG.md`

### 13.3 Дистрибуция

- Сборка: `goreleaser` → бинарники для Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Установка: `go install`, Homebrew tap, `.deb`/`.rpm` пакеты, Docker-образ

***

## 14. Приоритизация разработки

### MVP (Phase 1)
1. `api.Client` с авторизацией и обработкой ошибок
2. `singctl task` — полный CRUD
3. `singctl config` — управление токеном
4. Форматы вывода: table, json

### Phase 2
5. `singctl project`, `singctl habit`, `singctl tag`, `singctl time`
6. TUI: базовый вид задач и проектов
7. Shell completions

### Phase 3
8. TUI: все разделы + канбан-вид
9. `quick-add` синтаксис
10. Локальный кэш
11. Goreleaser + дистрибуция

---

## References

1. [API Server and Base Path | Swagger Docs](https://swagger.io/docs/specification/v3_0/api-host-and-base-path/)

2. [Wiki. FAQ: Public resources - SingularityApp](https://singularity-app.com/wiki/faq-public-resources/) - Is there a public API? Yes! The public API is available in the Pro and Elite plans. You can learn mo...
