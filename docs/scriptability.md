# Scriptability: коды выхода, потоки и pipe

Контракт CLI `singctl` для скриптов и CI (ТЗ §10, constitution V).
Числовой SoT в runtime: `cli.ExitCode` (см. также `specs/005-error-retry/contracts/cli-exit-codes.md`).

## Коды выхода

| Код | Смысл |
|-----|--------|
| `0` | Успех |
| `1` | Ошибка API, операции, транспорта или **использования CLI** (неизвестная команда/флаг, неверное значение флага) |
| `2` | Ошибка конфигурации (например, нет токена) |
| `3` | Сущность не найдена |

Краткая сводка также выводится в `singctl --help`.

## stdout и stderr

На обычном (не debug) пути:

| Исход | stdout | stderr |
|-------|--------|--------|
| Успех | Полезные данные / сообщение успеха | Пусто |
| Ошибка | Пусто | Сообщение об ошибке |
| `--help` / `--version` | Текст справки/версии | Пусто |

## Pipe-сценарии (ТЗ §10)

Полный live E2E с `task`/`time` — в F08+. Ниже — контрактные id и статус.

| ID | Пример | Статус |
|----|--------|--------|
| `json-redirect` | `singctl task list --output json > tasks.json` | `verifiable_now` (формат/ANSI через `internal/output` + streams); команда list — F08+ |
| `list-jq-xargs` | list JSON → `jq '.[].id'` → следующая команда | `contract_f08_plus` |
| `csv-awk` | `singctl time list --output csv \| awk …` | `verifiable_now` (CSV fixture F06); `time list` — F08+ |
| `xargs-create` | `xargs … singctl task create --title "{}"` | `contract_f08_plus` (независимый exit на каждый вызов; stdin-body create вне F07) |

F08: примеры `task list` / `task create` выше проверяются unit/httptest harness’ом (`internal/cli/task_*_test.go`); live pipe E2E — F33.
F09: `task checklist list` / `add` — тот же harness (`internal/cli/task_checklist_*_test.go`); pre-check задачи на list/add.
F10: `task kanban list` / `task move` — harness (`internal/cli/task_kanban_*_test.go`, `task_move_test.go`); list без pre-check; create/move — pre-check задачи.

Общие свойства пайплайнов: при успехе exit `0` и данные только в stdout; при ошибке ненулевой код (`1`/`2`/`3`) и сообщение в stderr; в pipe нет ANSI (F06); `json`/`yaml` list — корневой массив; одна задача (get/create/…) — один объект (F08 SingleObject).

Подробная матрица: `specs/007-scriptability-exits/contracts/pipe-scenarios.md`.
