# Contract: Pipe Scenarios from TZ §10 (F07)

**Feature**: `007-scriptability-exits` | **Date**: 2026-07-16
**Source**: `docs/tz/singularityapp-cli-tui-tz.md` §10

Каждый пример MUST иметь статус. Entity CRUD E2E не входит в DoD F07.

---

## Shared properties (all scenarios)

| Property | Requirement |
|----------|-------------|
| Exit on success | `0` |
| Exit on failure | `1` / `2` / `3` per [exit-codes-public.md](./exit-codes-public.md) |
| Data vs errors | [stream-separation.md](./stream-separation.md) |
| Color in pipe | No ANSI on non-TTY stdout (F06) |
| Machine formats | `json`/`yaml`/`csv` per F06 (`json`/`yaml` root array of objects) |

---

## Scenario matrix

### 1. `json-redirect` — list → JSON file

```bash
singctl task list --output json > tasks.json
```

| Property | Value |
|----------|-------|
| Format | `--output json` (or config) |
| stdout | Valid JSON array; no ANSI when redirected |
| stderr | Empty on success |
| Status | **contract_f08_plus** for `task list`; **verifiable_now** via `internal/output` JSON fixture + stream rules on `config show -o json` |
| Notes | F08+ supplies entity list command |

### 2. `list-jq-xargs` — list → ids → next command

```bash
singctl task list --project P-123 --output json \
  | jq -r '.[].id' \
  | xargs -I{} singctl task archive {}
```

| Property | Value |
|----------|-------|
| JSON shape | Root array; objects with stable `id` field (entity schema F08+) |
| Intermediate failure | Non-zero exit; message on stderr; no false success payload |
| Status | **contract_f08_plus** (needs `task list` / `task archive`) |
| Verifiable now | F06 root-array contract; F07 exit/stream rules |

### 3. `csv-awk` — CSV → aggregation

```bash
singctl time list --from … --to … --output csv \
  | awk -F, 'NR>1 {sum+=$4} END {print sum/3600 " hours"}'
```

| Property | Value |
|----------|-------|
| Format | CSV with header row; stable column count; proper escaping |
| stdout | No ANSI when piped |
| Status | **contract_f08_plus** for `time list`; **verifiable_now** via F06 CSV fixture tests |
| Notes | Column index `$4` is entity-specific (F08+); F07 only guarantees CSV contract shape |

### 4. `xargs-create` — lines → repeated create via args

```bash
cat tasks.txt | xargs -I{} singctl task create --title "{}"
```

| Property | Value |
|----------|-------|
| Input | CLI args per invocation (not required JSON-stdin body in F07) |
| Exit | Independent per process; table 0/1/2/3 |
| Status | **contract_f08_plus** (`task create`); **verifiable_now** — documented expectation + misuse/exit harness |
| Notes | F07 does **not** add stdin body parser for create |

---

## Coverage checklist (SC-004)

- [x] json-redirect — documented + partial verify path
- [x] list-jq-xargs — documented as F08+
- [x] csv-awk — documented + F06 CSV verify path
- [x] xargs-create — documented as F08+

Implement phase: keep this file authoritative; user-facing summary may copy into `docs/scriptability.md`.
