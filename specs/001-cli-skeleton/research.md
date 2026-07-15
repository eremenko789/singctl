# Research: CLI Skeleton (F01)

**Feature**: `001-cli-skeleton` | **Date**: 2026-07-15

Все пункты Technical Context, требовавшие решений, закрыты ниже. NEEDS CLARIFICATION не осталось.

---

## 1. Go module path и toolchain

**Decision**: Модуль `github.com/eremenko789/singctl`; Go toolchain — актуальный stable (≥ 1.22). Entrypoint сборки — `./cmd/singctl` (как в `Makefile`).

**Rationale**: Remote `origin` уже указывает на этот путь; Constitution II требует один Go-бинарник; Makefile уже ожидает `go build -o bin/singctl ./cmd/singctl`.

**Alternatives considered**:
- Корневой `main.go` — расходится с Makefile и ТЗ §4 (`cmd/singctl/`).
- Другой module path (`singctl`, `sa-cli`) — хуже для `go get`/import и не совпадает с GitHub.

---

## 2. CLI framework: Cobra + Viper

**Decision**: Корневое дерево команд на `spf13/cobra`. Глобальные флаги — `PersistentFlags` на root. Viper подключается с привязкой persistent-флагов (`BindPFlag`), **без** `ReadInConfig` в F01.

**Rationale**: Constitution Technology Constraints и F01 backlog явно задают cobra/viper. Ранняя привязка флагов к Viper упрощает F02 (резолвинг конфига) без смены контракта CLI.

**Alternatives considered**:
- Только Cobra без Viper — проще сейчас, но ломает заявленный стек и усложняет F02.
- `urfave/cli` / `kong` — вне constitution.

---

## 3. Расположение команд: `internal/cli` vs `cmd/`

**Decision**: `cmd/singctl/main.go` только вызывает `cli.Execute()`; дерево команд в `internal/cli`.

**Rationale**: Тестируемый пакет без `main`; entity-команды позже добавятся как соседние файлы/пакеты под `internal/cli` или `internal/cli/<resource>` без смены бинарной точки входа. Совместимо с Makefile.

**Alternatives considered**:
- Все команды в `cmd/*.go` с `package main` — хуже тестировать, раздувает entry package.
- Полное дерево `cmd/task|project|...` из ТЗ §4 — вне scope F01; отложено.

---

## 4. Версия: `version` и `--version`

**Decision**: Одинаковый текст вывода строится из `internal/buildinfo` (поля `Name=singctl`, `Version`, `Commit`, `Date`). Root: `cobra.Command.Version` + кастомный `SetVersionTemplate` (или `SetVersionFunc`) для `--version`. Отдельная subcommand `version` печатает ту же строку/блок. Dev-плейсхолдеры: `Version=dev`, `Commit=unknown`, `Date=unknown` при отсутствии ldflags.

**Rationale**: Spec Clarifications (Session 2026-07-15): оба пути с одинаковым смыслом; вывод включает имя, версию и метаданные сборки. Placeholders допустимы Assumptions.

**Alternatives considered**:
- Только `--version` без subcommand — не соответствует clarification Option A.
- Только subcommand — ломает привычный UX CI/`--version`.

---

## 5. Валидация `--output` раньше `--help` / version

**Decision**: Тип значения флага через `pflag.Value` (кастомный `OutputFormat`), допустимые: `table|json|yaml|csv`, default `table`. Ошибка `Set()` возникает на этапе разбора флагов **до** рендера help/version. Сообщение об ошибке — на русском, exit ≠ 0.

**Rationale**: Clarification: валидировать строго везде, включая help/version. Встроенный Cobra `String` + `PersistentPreRun` часто не успевает до `--help`.

**Alternatives considered**:
- Валидация только в `PersistentPreRun` — help может показаться раньше ошибки.
- Валидация только на entity-командах — нарушает FR-006 / SC-005.

---

## 6. Вызов без аргументов

**Decision**: У root `RunE` (не `Run`) возвращает ошибку на русском в духе «команда не указана» / «TUI ещё не реализован»; `SilenceUsage` по желанию, чтобы не дублировать длинный usage; exit ≠ 0. `cobra.ExactArgs`/отсутствие подкоманды — штатный путь Cobra + явный RunE.

**Rationale**: Spec: Option B — ошибка, не TUI. Целевое «без args → TUI» отложено до F18.

**Alternatives considered**:
- Печатать help и exit 0 — противоречит FR-013 / SC-006.
- Заглушка TUI — out of scope.

---

## 7. Русскоязычный UX

**Decision**: `Short`/`Long` root и `version`, `Usage` строк флагов и тексты ошибок валидации/«нет команды» — на русском. Имена флагов и enum форматов — латиница по ТЗ. Где стандартные англоязычные сообщения Cobra прорываются (unknown command), перехватывать через `SetFlagErrorFunc` / обработку `Execute` ошибки и переводить/заменять ключевые случаи; минимум — покрыть FR-014 для help, описаний флагов и ошибок парсера из наших проверок.

**Rationale**: Clarification Option B; Assumptions: машинные идентификаторы остаются латинскими.

**Alternatives considered**:
- Английский по умолчанию Cobra — не проходит FR-014.
- Полная i18n-библиотека — избыточно для F01.

---

## 8. Сеть и out-of-scope команды

**Decision**: В F01 не добавлять зависимости/импорты HTTP-клиентов и не регистрировать `task|project|habit|tag|time|tui`. Тесты контракта проверяют отсутствие этих имён в `singctl --help`.

**Rationale**: FR-010, FR-011, constitution III/VI для этого релиза фичи.

**Alternatives considered**: Заглушки entity-команд с «not implemented» — отклонены spec (не регистрировать).

---

## 9. Сборка метаданных

**Decision**: Переменные в `internal/buildinfo` с возможностью `-ldflags "-X ...Version=... -X ...Commit=... -X ...Date=..."`. В F01 достаточно placeholders; `goreleaser` — F36. Опционально позже расширить `make build` ldflags — не блокер скелета.

**Rationale**: SC-002 требует наличия полей; Assumptions допускают `unknown`/`dev`.

**Alternatives considered**: Только runtime чтения VCS через `debug.ReadBuildInfo` — можно как дополнение, но ldflags привычнее для релизов.

---

## 10. Тестирование

**Decision**: Table-driven тесты вокруг `Execute` с подменой `os.Args` / `SetArgs`, буферами stdout/stderr. Сценарии: help (русский, глобальные флаги, нет entity), version vs `--version` (равенство), все валидные глобальные флаги, невалидный `--output` с help/version, bare invoke exit ≠ 0.

**Rationale**: Quality Gates constitution; независимые тесты user stories; без сети.

**Alternatives considered**: Только ручной smoke — недостаточно для SC-003/SC-005 автоматизации.
