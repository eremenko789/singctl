# Quickstart: validate CLI Skeleton (F01)

**Feature**: `001-cli-skeleton` | **Date**: 2026-07-15

Руководство для проверки end-to-end после реализации (см. [contracts/cli.md](./contracts/cli.md), [data-model.md](./data-model.md)). Без сетевого доступа.

---

## Prerequisites

- Go toolchain (stable) установлен
- Репозиторий с `go.mod` и исходниками F01 (после `/speckit-implement`)
- Команды из корня репозитория

---

## Setup

```bash
# при необходимости
cp .env.example .env   # для F01 не обязателен

make build
# ожидается: bin/singctl
```

Проверка: файл `bin/singctl` существует и исполняемый.

---

## Validation scenarios

### 1. Help (SC-001, FR-002, FR-014)

```bash
./bin/singctl --help
```

**Ожидание**:
- exit 0, < 2 s
- русский текст описания и флагов
- видны глобальные флаги из [contracts/cli.md](./contracts/cli.md)
- нет команд `task`, `project`, `habit`, `tag`, `time`, `tui`

### 2. Version parity (SC-002, FR-003)

```bash
./bin/singctl version
./bin/singctl --version
```

**Ожидание**:
- оба exit 0
- одинаковый stdout: имя `singctl`, версия, commit и/или date (допустимы `dev`/`unknown`)
- без сетевых обращений

### 3. Global flags accepted (SC-003, FR-004…008)

```bash
./bin/singctl --config /tmp/example.yaml --token TOKEN \
  --output json --no-color --debug version

./bin/singctl -o yaml --help
```

**Ожидание**: нет ошибки «unknown flag»; version/help при валидных флагах успешны.

### 4. Invalid output blocks help/version (SC-005, FR-006)

```bash
./bin/singctl --output xml --help
./bin/singctl -o xml version
./bin/singctl --output xml --version
```

**Ожидание**: для каждого — exit ≠ 0, русмое сообщение валидации на русском, **без** полного help/version payload.

### 5. Bare invoke (SC-006, FR-013)

```bash
./bin/singctl
```

**Ожидание**: exit ≠ 0; сообщение об ошибке; TUI не открывается.

### 6. Unknown command

```bash
./bin/singctl nosuch
```

**Ожидание**: exit ≠ 0; ошибка; подсказка к `--help` желательна.

---

## Automated check (после появления тестов)

```bash
make test
```

**Ожидание**: unit/contract-тесты CLI зелёные; сценарии выше покрыты без сети.

---

## Out of scope for this quickstart

- Чтение реального конфига / запись токена (F02)
- HTTP к API
- Entity CRUD и TUI
