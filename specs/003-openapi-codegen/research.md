# Research: OpenAPI Codegen Pipeline (F03)

**Feature**: `003-openapi-codegen` | **Date**: 2026-07-16

Все пункты Technical Context и отложенные из clarify решения закрыты. NEEDS CLARIFICATION не осталось.

---
## 1. Конфиг `oapi-codegen` и layout клиента

**Decision**: Использовать `api/oapi-codegen.yaml` ровно в форме из `docs/openapi-codegen.md`:

```yaml
package: apiclient
generate:
  models: true
  client: true
output: internal/apiclient/client.gen.go
```

Один файл `client.gen.go` в пакете `apiclient`. Ручные файлы в `internal/apiclient/` в F03 не добавляются (адаптер — F04 в другом пакете, напр. `internal/api`).

**Rationale**: Совпадает с docs/ТЗ; минимальный конфиг; один артефакт для ревью/коммита.

**Alternatives considered**:
- Раздельные `models.gen.go` / `client.gen.go` — больше шума в PR без выгоды на 51 ops.
- Генерация server/chi stubs — не нужна CLI-клиенту.

---
## 2. Независимость Make-таргетов и роли JSON vs YAML

**Decision**:
- `openapi-fetch`, `api-coverage-check`, `generate` — **без** Make-prerequisites друг на друга (clarify Q2).
- `api-coverage-check` читает **`docs/api/openapi.json`**.
- `generate` читает **`docs/api/openapi.yaml`** (как текущий Makefile) + `api/oapi-codegen.yaml`.
- Рекомендуемый порядок только в docs / тексте `make help`: fetch → coverage-check → generate.

**Rationale**: Позволяет чинить снимок, ожидание ops и клиент по отдельности; соответствует clarify.

**Alternatives considered**:
- `generate` depends on `api-coverage-check` — отклонено (clarify B).
- Генерировать из JSON — возможно, но YAML уже канон в Makefile/docs.

---
## 3. Строгость `api-coverage-check`

**Decision**: Успех = (число HTTP operations в JSON == `EXPECTED_API_OPS`, default **51**) **и** существует `docs/api/coverage.md`. Без парсинга строк матрицы / operationId (clarify Q3).

Подсчёт: как сейчас — методы в `paths.*`, исключая ключи `x-*`.

**Rationale**: Дёшево, стабильно, достаточно для acceptance F03; глубокая сверка матрицы — F35 / позже.

**Alternatives considered**:
- Сверка числа строк markdown / operationId — отложено (шум форматирования).

---
## 4. Атомарность `openapi-fetch` (deferred из clarify → plan)

**Decision**: Реализовать fetch через временные файлы + `mv` (или эквивалент): сначала скачать JSON и YAML во временные пути в `docs/api/`, затем атомарно заменить целевые файлы **только если оба скачивания успешны**. При ошибке второго запроса целевые файлы не трогать (или не оставлять пару рассинхронизированной намеренно).

**Rationale**: Edge case «JSON обновлён, YAML нет» ломает рассинхрон coverage (JSON) vs generate (YAML). Атомарная пара закрывает риск без усложнения scope.

**Alternatives considered**:
- Оставить текущий прямой `curl -o` — проще, но допускает частичный апдейт.
- Rollback/git checkout при сбое — хрупко вне чистого tree.

---
## 5. Версия `oapi-codegen` и воспроизводимость

**Decision**: Установка разработчика — `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest` (как в docs). Source of truth для runtime API-типов в репозитории — **закоммиченный** `client.gen.go` (clarify Q1). Автоматический no-diff gate после generate — **не** в F03 (clarify Q4).

При заметном дрейфе вывода между машинами — зафиксировать точный тег в docs/`Makefile` комментарии отдельным follow-up; для F03 достаточно v2 `@latest` + коммит артефакта.

**Rationale**: Balance между простотой онбординга и DoD «клон собирается офлайн». Drift-gate отложен сознательно.

**Alternatives considered**:
- `tools.go` + pin exact version — полезно позже, не блокер F03.
- Не коммитить gen, только CI generate — отклонено (clarify A).

---
## 6. Тестирование и coverage gate для gen-кода

**Decision**:
- Контрактные проверки: `make api-coverage-check` (exit 0 на каноне; exit ≠ 0 при неверном `EXPECTED_API_OPS`); `make generate` (exit 0 при конфиге + инструменте); негатив — отсутствие конфига.
- После generate: пакет компилируется (`go build ./internal/apiclient/...` или через `make build` модуля).
- `*.gen.go` MAY исключаться из coverage profile / gate в `make test` (constitution IX) — настроить при касании Makefile test, если gen начинает портить метрики.
- Новый ручной Go-код (если появится хелпер) — только TDD; предпочтение — не раздувать F03 новым пакетом, оставить логику в Make/python one-liner.

**Rationale**: F03 — пайплайн, не бизнес-логика; gen не покрывается unit-тестами осмысленно.

---
## 7. Документация и DoD коммита

**Decision**: После реализации синхронизировать `docs/openapi-codegen.md` и `docs/makefile.md` с фактическим поведением (рекомендуемый порядок, строгость check, атомарный fetch). В git MUST: снимок, `coverage.md`, `api/oapi-codegen.yaml`, `internal/apiclient/*.gen.go`. MUST NOT: `.env`, токены.

**Rationale**: FR-010/012 + clarify Q1.

---
## 8. Итог

**Decision summary**:
- Single-file oapi-codegen client в `internal/apiclient/`.
- Независимые Make-таргеты; JSON→coverage, YAML→generate.
- Coverage-check = ops count + file exists.
- Atomic dual-file fetch.
- `@latest` v2 + committed gen; no drift-gate in F03.
- Contractual Make validation; exclude gen from coverage gate as needed.
