# Contract: API Errors & Retry (F05)

**Feature**: `005-error-retry` | **Date**: 2026-07-16
**Package**: `internal/api` (поверх F04 `HTTPError` / `Session`)

Имена символов — ориентир для implement; семантика обязательна.

---

## Classify

```text
Classify(err error, opts ...ClassifyOption) error
```

- `nil` → `nil`
- `*HTTPError` → `*ClassifiedError` with Kind + Message per catalog in [data-model.md](../data-model.md)
- Option `WithEntityID(id)` applies to 404 message when id non-empty
- Config / missing-token errors → Kind Config (preserves actionable set-token hint)
- `*DateError` → Kind Date/Validation with date hint
- Other (transport) → Kind Transport, no HTTP status

`ClassifiedError` MUST support `errors.As` / `Unwrap` to underlying cause when present.

**Non-goals**: TUI widgets; changing codegen types.

---

## EnsureSuccess (F04)

Unchanged: non-2xx → `*HTTPError`. Callers that need UX text SHOULD `Classify` the result.

---

## Retry transport

Session HTTP client MUST use a RoundTripper such that:

| Response / error | HTTP attempts | Notes |
|------------------|---------------|-------|
| 429 then 2xx within 3 | ≤ 3 | success |
| 429 × 3 | exactly 3 | then `HTTPError` 429 → Classify → `Error: rate limited. Retry later` |
| 401/403/404/422/5xx | exactly 1 | no retry |
| transport failure | 1 | no retry |
| ctx cancel during backoff | stop | return ctx error |

**Backoff**:
- Valid `Retry-After` (seconds or HTTP-date) → wait `min(duration, 30s)`
- Else → 1s before 2nd attempt, 2s before 3rd
- Tests MUST inject sleeper (no real multi-second sleeps)

**Request identity**: retries MUST repeat the same method/URL/headers/body semantics.

---

## ParseDate

```text
ParseDate(s string) (time.Time, error)
```

| Input | Result |
|-------|--------|
| `2025-11-28` | OK |
| `28.11.2025`, `2025/11/28`, `2025-13-01`, `""` | error containing `Expected: YYYY-MM-DD` |

No CLI command required in F05.

---

## ValidateConnectivity (update)

- Still uses `ProjectControllerListWithResponse`
- Benefits from session retry transport automatically
- On failure, prefer returning `Classify(err)` (list has no entity ID)
- Comment «exactly one HTTP attempt» from F04 MUST be updated: one logical call, up to 3 HTTP on 429

---

## Testing contract

- Table test: status → Kind + Message (401/403/404±ID/422/429/5xx)
- Retry: 429→429→200 (3 hits); 429×3 (3 hits + rate limited message); 404 (1 hit)
- `Retry-After: 1` respected (sleeper records duration); oversized capped at 30s
- ParseDate unit tests only
- Fixtures: `test-token-…` only

---

## Out of scope

- Retry on 5xx
- TUI banners
- Entity CRUD facades
