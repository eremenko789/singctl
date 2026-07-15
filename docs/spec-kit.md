# GitHub Spec Kit в этом репозитории

Проект ведётся по Spec-Driven Development с помощью структуры [GitHub Spec Kit](https://github.com/github/spec-kit).

## Карта артефактов

```text
.specify/
├── memory/constitution.md   # Неизменяемые принципы проекта
└── templates/               # Шаблоны Spec Kit (reference)

docs/
├── tz/                      # Исходное ТЗ (вход)
├── api/                     # OpenAPI + wiki snapshot
├── openapi-codegen.md       # Повторяемые операции codegen
└── spec-kit.md              # Этот файл

specs/001-singctl-cli-tui/
├── spec.md                  # WHAT: user stories + requirements
├── plan.md                  # HOW: стек, архитектура
├── research.md              # Решения Phase 0
├── data-model.md            # Сущности
├── quickstart.md            # Сценарии проверки
├── contracts/openapi.yaml   # Контракт API для фичи
├── checklists/requirements.md
└── tasks.md                 # Исполняемые задачи
```

## Workflow

1. **Constitution** — принципы в `.specify/memory/constitution.md`.
2. **Specify** — `specs/.../spec.md` (что строить).
3. **Plan** — `plan.md` + `research.md` + `data-model.md` + `contracts/` + `quickstart.md`.
4. **Tasks** — `tasks.md`.
5. **Implement** — реализация по задачам (отдельные PR).

Исходное ТЗ (`docs/tz/`) сохраняется как исторический вход; рабочие артефакты для агентов и разработки — в `specs/`.

## Именование фич

Каталоги: `specs/NNN-short-name/` (например `001-singctl-cli-tui`). Новые фичи — следующие номера.
