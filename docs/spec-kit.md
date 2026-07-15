# GitHub Spec Kit в этом репозитории

Проект готовится к Spec-Driven Development по [GitHub Spec Kit](https://github.com/github/spec-kit).

## Что уже есть (отправные материалы)

Это **не** результаты `/speckit.specify|plan|tasks`, а вход и bootstrap:

```text
.specify/
├── memory/constitution.md   # принципы (в т.ч. Makefile/.env, OpenAPI codegen)
└── templates/               # пустые шаблоны Spec Kit (как после specify init)

docs/
├── tz/                      # исходное ТЗ (вход для specify)
├── api/                     # OpenAPI-снимок + coverage matrix + wiki
├── openapi-codegen.md
├── makefile.md
└── spec-kit.md              # этот файл

Makefile
.env.example
```

## Чего здесь специально нет

Каталог `specs/<NNN>-*/` с заполненными `spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`, `tasks.md`, checklists **не создаётся вручную**. Его порождает Spec Kit:

| Команда | Результат |
|---|---|
| `/speckit.constitution` | уточнение `.specify/memory/constitution.md` |
| `/speckit.specify` | `specs/<NNN>-name/spec.md` |
| `/speckit.clarify` | правки spec |
| `/speckit.plan` | `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md` |
| `/speckit.checklist` | `checklists/` |
| `/speckit.tasks` | `tasks.md` |
| `/speckit.implement` | код |

Рекомендуемый поток:  
`constitution → specify → clarify → plan → checklist → tasks → analyze → implement → converge`  
(см. [quickstart Spec Kit](https://github.github.com/spec-kit/quickstart.html)).

## Вход для `/speckit.specify`

Указывать агенту:

- продуктовое ТЗ: `docs/tz/singularityapp-cli-tui-tz.md`
- полное покрытие API: `docs/api/coverage.md` (все 51 operations)
- OpenAPI: `docs/api/openapi.yaml`
- стек и ограничения: `.specify/memory/constitution.md` (Go, codegen, Makefile)

## Именование фич

После specify каталоги будут вида `specs/NNN-short-name/`.
