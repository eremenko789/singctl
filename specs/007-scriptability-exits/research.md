# Research: Scriptability & Exit Codes (F07)

**Feature**: `007-scriptability-exits` | **Date**: 2026-07-16

Все пункты Technical Context и clarify-сессии закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Где жить пользовательская документация кодов выхода

**Decision**: Новый файл **`docs/scriptability.md`** (полная таблица `0|1|2|3`, правила stdout/stderr, обзор pipe). Строка в указателе **`docs/README.md`**. Спек-контракты в `specs/007-…/contracts/` остаются для implement/review, не заменяют user docs.

**Rationale**: Clarify Q1 — полная таблица в `docs/`, обнаруживаемость из репо; стиль проекта (остальные user docs в `docs/`). Constitution Documentation Expectations.

**Alternatives considered**:
- Только `specs/…/contracts/` — не user-facing (отклонено clarify).
- Только README корня репо — смешивает install/dev и CLI-контракт.
- Разнести exit vs pipe по двум md — избыточно для объёма F07; один файл проще для SC-001.

---

## 2. Форма упоминания в корневом `--help`

**Decision**: Расширить `Long` (или добавить `Example`) у root cobra-команды краткой сводкой:

```text
Коды выхода: 0 успех; 1 ошибка API/операции/использования; 2 конфигурация; 3 не найдено.
Подробнее: docs/scriptability.md
```

Тест: `executeForTest([]string{"--help"})` — stdout содержит `0`/`1`/`2`/`3` (или явные слова «выход»/«exit») и не падает.

**Rationale**: Clarify Q1 — бинарь обнаруживает контракт; не дублировать весь docs в help.

**Alternatives considered**:
- Отдельная команда `singctl help exit-codes` — новая команда, вне scope.
- Только URL без таблицы — хуже offline/скрипты.
- Печать кодов в Version — нестандартно.

---

## 3. Misuse CLI → exit 1

**Decision**: Не вводить Kind Config для flag/command errors. Неизвестный флаг/команда, неверный `--output` остаются обычными `error` → `ExitCode` → **`1`** (как сейчас для non-Classified). Явный тест: `--unknown-flag` / `--output xml` → `ExitCode(err)==1`, stdout пуст, stderr непуст.

**Rationale**: Clarify Q2; ТЗ §10 резервирует `2` за конфигурацией (токен/файл настроек), не за опечаткой во флаге. F05 `ExitCode` уже так себя ведёт.

**Alternatives considered**:
- Map misuse → 2 — путает с «нет токена».
- Отдельный exit 64 (sysexits) — ломает таблицу ТЗ из четырёх кодов.

---

## 4. Пустой stdout при ошибке / пустой stderr при успехе

**Decision**: На DoD-поверхности (`config show`, `config validate`, misuse) ужесточить assertions:

| Исход | stdout | stderr |
|-------|--------|--------|
| Успех (не debug) | data / success message | **пусто** (после trim) |
| Ошибка | **пусто** | user-facing сообщение (+ префикс `Ошибка:` из Execute) |

`cli.Execute` / `executeForTest` уже пишут ошибки в stderr. Если какой-то путь пишет успех/ошибку не туда — минимальный фикс команды, не новый фреймворк.

`--help` / `--version`: успех с текстом в **stdout**, stderr пуст — допустимо (help/version = полезный вывод, не «диагностика»).

**Rationale**: Clarify Q3/Q4; pipe-safe. Текущие validate-тесты местами игнорируют stdout на ошибке (`_ , stderr`) — F07 закрывает дыру.

**Alternatives considered**:
- Буферизовать весь stdout до commit-on-success — избыточно без стриминга; для текущих команд достаточно «не писать data до успеха» / не писать ошибку в Out.
- Разрешить warnings в stderr на успехе — отклонено clarify.

---

## 5. Контракт pipe-сценариев ТЗ §10 без entity-команд

**Decision**: Файл `contracts/pipe-scenarios.md` с четырьмя примерами; колонки: свойства (format, ANSI, streams, exit) + статус **Verifiable now** (F06 fixture / config) vs **Contract for F08+** (task/time CRUD).

Автопроверка «сейчас»:
- JSON/CSV/ANSI — reuse `internal/output` tests (F06);
- streams/exit — show/validate/misuse (F07);
- list→jq→archive / xargs create — только контрактная запись до F08+.

**Rationale**: Acceptance «на уровне контракта»; clarify Q5 / FR-005/011; Honest API Boundaries (G6).

**Alternatives considered**:
- Скрытая demo `task list` — новая команда, вне scope.
- Отложить весь §10 до F08 — ломает acceptance F07.

---

## 6. Non-interactive / stdin pipe

**Decision**: Для существующих команд payload из stdin не читается. Тест: `cmd.SetIn` с закрытым/finite reader + `config show` или `version` — завершение без hang (timeout в тесте). Ошибки сразу в stderr + ненулевой ExitCode, без confirm UI (его пока нет).

**Rationale**: Spec US4 / FR-007; backlog «stdin/pipe».

**Alternatives considered**:
- JSON-stdin API для create в F07 — нет create-команды.
- Детект stdin TTY для авто-режимов — не требуется ТЗ сверх color-по-stdout (F06).

---

## 7. Связь с F05 ExitCode (единственный SoT)

**Decision**: Не дублировать mapping в новом helper. `docs/scriptability.md` и `contracts/exit-codes-public.md` **описывают** ту же таблицу, что `cli.ExitCode` / F05 `cli-exit-codes.md`. F07 может добавить cross-link в godoc `ExitCode`.

**Rationale**: FR-010; avoid competing tables.

**Alternatives considered**:
- Переписать ExitCode в F07 — ненужный churn.
- Вынести коды в отдельный пакет — overkill.

---

## 8. Язык документации

**Decision**: **`docs/scriptability.md` на русском** (как `docs/README.md` и UX CLI), с английскими стабильными фрагментами catalog errors где цитируется F05 (`Error: entity not found`). Коды и смысл однозначны в обоих языках help-сводки.

**Rationale**: Spec assumption MAY RU/EN; консистентность репо.

**Alternatives considered**:
- Только EN docs — расхождение с остальным `docs/`.
