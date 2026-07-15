# sa-cli (singctl) Constitution

## Core Principles

### I. Spec-Driven Development

Проект следует [GitHub Spec Kit](https://github.com/github/spec-kit). Артефакты фич (`specs/<NNN>-*/spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`, `tasks.md`, checklists) создаются **только** командами Spec Kit (`/speckit.specify`, `/speckit.plan`, `/speckit.tasks`, …), а не вручную заранее.

Входные материалы для Specify живут в `docs/` (ТЗ, OpenAPI-снимок, матрица покрытия API). Код следует актуальной спецификации в `specs/`; при изменении требований сначала обновляется spec/plan/tasks через Spec Kit, затем код.

### II. Go Single Binary

Язык реализации — Go. Цель поставки — один статически линкуемый CLI/TUI-бинарник (`singctl`) без обязательного рантайма. Зависимости минимальны и обоснованы.

### III. OpenAPI-Generated API Client (NON-NEGOTIABLE)

HTTP-обёртки вокруг SingularityApp REST API **не пишутся вручную**. Клиент и модели генерируются из OpenAPI (`docs/api/openapi.yaml` / upstream `https://api.singularity-app.com/v2/api-json`).

Ручной код допускается только как тонкий адаптер поверх сгенерированного клиента (auth, retry, mapping ошибок, CLI/TUI-фасады). Повторная генерация — через Makefile (см. принцип VIII) и `docs/openapi-codegen.md`.

Клиент MUST покрывать **все** операции публичного REST API v2 из актуального OpenAPI (см. `docs/api/coverage.md`). Пропуск endpoint без явной фиксации out-of-scope в Spec Kit spec запрещён.

### IV. Shared Client for CLI and TUI

CLI (`cobra`) и TUI (`bubbletea`) используют один и тот же API-слой. Бизнес-логика не дублируется в UI.

### V. Scriptability First

CLI должен быть удобен для pipe/скриптов: стабильные exit codes, форматы `table|json|yaml|csv`, авто-отключение цвета в non-TTY, предсказуемый stdout/stderr.

### VI. Honest API Boundaries

Нельзя обещать функции, которых нет в API (webhooks/SSE, создание повторяющихся задач, совместные проекты, офлайн-синхронизация). Ограничения явно отражаются в UX и документации.

### VII. Security of Credentials

Токен API хранится только в локальном конфиге пользователя, маскируется в `config show` и логах. Токен не коммитится в репозиторий. Секреты и локальные параметры разработчика — в `.env` (не в git); в debug-логах секреты редактируются.

### VIII. Makefile + `.env` for Recurring Work (NON-NEGOTIABLE)

Регулярные задачи разработки **документируются и автоматизируются через `Makefile`**. Цель Make — единая точка входа вместо «магических» shell one-liner’ов в чатах.

Правила:

- Повторяемые операции (fetch OpenAPI, codegen, build, test, lint, smoke против API, release prep) имеют именованные `make`-таргеты.
- Параметры окружения (токен, base URL, флаги CI и т.п.) при необходимости читаются из файла `.env` в корне репозитория (`include .env` / `export` / явная подстановка). Пример — `.env.example` (без секретов).
- `.env` **никогда** не коммитится; в репозитории только `.env.example` с пустыми/фиктивными значениями.
- Длинное описание «зачем/когда» остаётся в `docs/`; Make-таргеты — канонический способ **запуска**.
- Новая повторяемая процедура: сначала таргет в `Makefile` (+ строка в `.env.example` при необходимости), затем краткая ссылка из `docs/`.

## Technology Constraints

| Область | Выбор |
|---|---|
| Язык | Go (актуальный stable) |
| CLI | `spf13/cobra` + `spf13/viper` |
| TUI | `charmbracelet/bubbletea`, `lipgloss`, `bubbles` |
| API client | генерация из OpenAPI (`oapi-codegen` или эквивалент) |
| HTTP | stdlib `net/http` / транспорт сгенерированного клиента |
| Вывод таблиц | `olekukonko/tablewriter` (или совместимый) |
| Автоматизация | `Makefile` + `.env` / `.env.example` |
| Релизы | `goreleaser` (поздняя фаза) |

Альтернативы (Python и т.п.) из исходного ТЗ **не используются** в этом репозитории.

## Quality Gates

- Unit-тесты для адаптеров API (мок HTTP) и парсеров CLI.
- TUI: model-тесты на `tea.Msg` там, где это практично.
- Integration-тесты с реальным API — только при наличии тестового токена (из `.env`) и явном включении.
- Каждая команда имеет `--help`.
- Перед реализацией фичи: Spec Kit-цепочка constitution → specify → plan → tasks (артефакты в `specs/` появляются только после этих команд).
- Покрытие CLI/TUI сверяется с `docs/api/coverage.md` (все операции OpenAPI).

## Documentation Expectations

Повторяемые операции описываются в `docs/` и запускаются через `Makefile`. README ссылается на docs и Make-таргеты, не дублирует длинные процедуры.

## Governance

Constitution имеет приоритет над ad-hoc решениями в коде и чатах. Изменения принципов фиксируются в этом файле с обновлением версии и даты.

**Version**: 1.1.0 | **Ratified**: 2026-07-15 | **Last Amended**: 2026-07-15
