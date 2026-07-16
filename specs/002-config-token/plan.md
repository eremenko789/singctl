# Implementation Plan: Config & Token Storage (F02)

**Branch**: `002-config-token` | **Date**: 2026-07-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-config-token/spec.md`

Note: Этот файл заполняет workflow `/speckit-plan` для F02. Он фиксирует дизайн/контракты и ожидаемую структуру реализации. Код и полные тест-сьюты создаются на фазе `/speckit-tasks` и `/speckit-implement`.

---
## Summary

F02 добавляет группу команд `singctl config`:
- `set-token <TOKEN>`: сохраняет токен в локальный YAML-конфиг в `api.token` (без префикса `Bearer` в файле);
- `show`: выводит effective конфигурацию, включая замаскированный токен, и уважает глобальный `--output`/`-o`;
- `set <key> <value>`: обновляет допустимые ключи конфигурации (dotted path) согласно схеме из ТЗ §5.1;
- `validate`: проверяет готовность к работе с API как честная локальная заглушка до появления полного HTTP/API слоя (и не выдаёт ложный "OK").

CRUD сущности и сетевые вызовы вне `validate` не входят в F02.

---
## Technical Context

**Language/Stack**: Go (constitution II). CLI на базе `spf13/cobra` + `spf13/viper` (конституция/стек проекта).

**Storage**:
- YAML-конфиг пользователя в одном из источников: `--config`, локальный `./.singctl.yaml`, XDG, `~/.config`.
- `set-token` и `set` создают/обновляют файл в резолвленном пути (с созданием каталогов при необходимости).

**Token handling**:
- в файле хранится "голый" токен (`api.token`);
- перед отправкой в Authorization заголовок добавляется `Bearer` (с пробелом после).

**Output behavior**:
- `config show` уважает `--output`/`-o` (`table|json|yaml|csv`);
- токен всегда маскируется (первые 4 + `****` + последние 4; короткие — `****`).

**Validate behavior**:
- до появления полноценного API слоя `validate` не выполняет CRUD и не обещает удалённую проверку;
- если токена нет — команда завершает ошибкой с подсказкой `set-token`.

**Testing**:
- unit/contract tests парсеров команд и валидации (TDD, constitution IX);
- тесты чтения/записи конфигурации с временной директорией и фиктивным `XDG_CONFIG_HOME` (без реального API).

---
## Constitution Check

| Gate | Principle | Status | Notes |
|------|-----------|--------|-------|
| G1 | Spec-Driven Development | PASS | план/research/data-model/contracts/quickstart созданы по spec |
| G2 | Go Single Binary | PASS | влияние только на Go-CLI слой |
| G3 | OpenAPI-Generated API Client | PASS (N/A) | validate допускает stub до появления клиента |
| G4 | Shared Client for CLI and TUI | PASS (N/A) | TUI отсутствует в F02 |
| G5 | Scriptability First | PASS | `show -o` / exit codes по контракту |
| G6 | Honest API Boundaries | PASS | `validate` честная заглушка до появления API |
| G7 | Security of Credentials | PASS | токен "голый" в файле; маскирование в show |
| G8 | Makefile + .env | PASS (N/A) | F02 не требует .env в первом проходе |
| G9 | TDD & Coverage | PASS | tests будут добавлены на фазе tasks/implement |

---
## Project Structure (design intent)

Пакеты/модули, ожидаемые для F02 (не являются реализацией кода):

1. `internal/config`
   - резолвинг effective config path (`--config` → `./.singctl.yaml` → XDG → `~/.config`);
   - загрузка/сохранение YAML;
   - валидация схемы `ConfigDocument` и допустимых ключей для `config set`;
   - правила нормализации токена и маскирования для `config show`.

2. `internal/cli/config`
   - реализация subcommands `set-token`, `show`, `validate`, `set`;
   - интеграция с root global options (`--output`/`-o`, `--token`, `--no-color`, `--debug`).

3. Общие helpers (по проектным соглашениям)
   - форматирование вывода (табличный/JSON/YAML/CSV) для `config show` согласно `--output`;
   - безопасные сообщения об ошибках, исключающие печать токена.

---
## Phase 0: Outline & Research (Output = research.md)

Решения зафиксированы в `research.md`:
- нормализация токена "голый" <-> `Bearer` (с пробелом после) в Authorization;
- маскирование в `config show`;
- приоритет резолвинга конфиг-пути;
- схема и допустимые ключи для `config set`;
- stub-поведение `config validate` до API слоя.

---
## Phase 1: Design & Contracts (Output = data-model.md, contracts/*, quickstart.md)

Artifacts:
- `data-model.md`: формальная модель конфиг-документа и effective-настроек
- `contracts/cli-config.md`: контракт поведения `singctl config`
- `quickstart.md`: проверка end-to-end UX в изолированном `XDG_CONFIG_HOME`

---
## Next Phase (not executed here)

Дальше `/speckit-tasks` должен:
- разложить реализацию на TDD-задачи (tests до/вместе с кодом) с учётом constitution IX;
- установить зависимости: `config` reading/writing -> show/set-token -> validate -> output formats & error paths.
