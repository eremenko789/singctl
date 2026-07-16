# Research: API Adapter & Auth (F04)

**Feature**: `004-api-adapter-auth` | **Date**: 2026-07-16

Все пункты Technical Context и спорные места закрыты. NEEDS CLARIFICATION не осталось.

---

## 1. Поверхность адаптера vs codegen

**Decision**: Пакет `internal/api` предоставляет фабрику `Session`, держащую `*apiclient.ClientWithResponses`, плюс хелперы маппинга ошибок и один probe-метод для validate/happy path. Полные фасады сущностей не добавляются.

**Rationale**: Clarification F04 (вариант B) + constitution III («тонкий адаптер»: auth, mapping; фасады CLI/TUI — позже).

**Alternatives considered**:
- Только фабрика без маппинга — ломает контракт типизированной ошибки для F05.
- Фасад всех 51 операций — out of scope и дублирует codegen.

---

## 2. Auth: Bearer через RequestEditor

**Decision**: При создании клиента регистрировать `apiclient.WithRequestEditorFn`, который выставляет `Authorization: Bearer <token>`. Токен из конфига — «голый»; если уже начинается с `Bearer ` (с пробелом), префикс не дублировать.

**Rationale**: Согласовано с F02 FR-004a; oapi-codegen официально поддерживает `WithRequestEditorFn` для auth.

**Alternatives considered**:
- Передавать editor на каждый вызов — легко забыть; фабрика гарантирует единообразие.
- Кастомный `http.RoundTripper` — избыточно при наличии RequestEditor.

---

## 3. Base URL и paths `/v2/...`

**Decision**: `api.base_url` — origin хоста API **без** суффикса `/v2` (например `https://api.singularity-app.ru`). Codegen склеивает `server` + `./v2/project` (относительный path); `NewClient` сам добавляет trailing `/`.

**Rationale**: В снимке OpenAPI `servers: []`, а operation paths уже `/v2/...`. Значение default из F02 совпадает. Суффикс `/v2` в base URL дал бы `/v2/v2/...`.

**Alternatives considered**:
- Нормализовать/стрипать `/v2` в адаптере — можно как защиту позже; в F04 достаточно документировать контракт и fail на пустой URL.
- Менять codegen/OpenAPI servers — вне scope F04.

---

## 4. Timeout

**Decision**: Парсить `api.timeout` (`time.ParseDuration`) и передавать `apiclient.WithHTTPClient(&http.Client{Timeout: d})`. Невалидная длительность / пустой timeout после defaults → ошибка фабрики.

**Rationale**: Один session-wide timeout покрывает FR-004 без per-call context plumbing в каждом будущем фасаде; context от вызывающего всё ещё может отменить раньше.

**Alternatives considered**:
- Только `context.WithTimeout` на каждом вызове — удобно точечно, но легко забыть в CLI; session default нужен всё равно.
- Transport-level deadlines — сложнее, без выигрыша в F04.

---

## 5. Типизированная ошибка не-2xx

**Decision**: Тип `api.HTTPError` с полями минимум `StatusCode int`, `Body []byte` (опционально усечённое сообщение для `Error()`). Метод `StatusCode()` / экспортированное поле — для `errors.As` в F05. Хелпер `EnsureSuccess(status int, body []byte) error` (или аналог по response codegen) возвращает `nil` на 2xx и `*HTTPError` иначе. Без классификации 401/429/… и без retry.

**Rationale**: Clarification F04 (вариант B); F05 нарастит taxonomy поверх стабильного типа.

**Alternatives considered**:
- Ошибка только строкой — хрупко для F05.
- Полная taxonomy в F04 — scope creep в F05.

---

## 6. Репрезентативная операция (happy path + validate)

**Decision**: Canonical probe — `ProjectController_list` / `ProjectControllerListWithResponse` (`GET /v2/project`). Параметры list: `nil` или пустые (без обязательных query). Успех: HTTP 2xx и возможность прочитать статус; тело JSON200 при 200 используется в unit-тесте при наличии валидной фикстуры.

**Rationale**: Стабильный read/list без path id; не создаёт сущностей; подходит для «лёгкой» проверки connectivity в `config validate`.

**Alternatives considered**:
- `TagController_list` — тоже ок, но project — базовая сущность ТЗ.
- `TaskController_list` — часто с фильтрами; тяжелее для минимального мока.
- Отдельный health endpoint — в публичном REST v2 нет.

---

## 7. Fail-fast фабрики

**Decision**: `NewSession` / `NewFromSettings` возвращает ошибку (русское сообщение, без токена в тексте), если token пуст/whitespace или base URL пуст/whitespace. Сеанс не создаётся. Невалидный timeout — тоже ошибка фабрики.

**Rationale**: Clarification F04; упрощает `config validate` без токена.

---

## 8. Интеграция `config validate`

**Decision**: Команда загружает effective settings → `api.NewFromSettings` → `Session.ValidateConnectivity(ctx)`. Успех: exit 0 + сообщение об успешной **удалённой** проверке (не «заглушка»). Ошибки фабрики/HTTP/transport → ненулевой код, без утверждения OK. В CLI-тестах: `httptest.Server` + `api.base_url` = URL сервера в временном конфиге.

**Rationale**: DoD clarification; тот же сеанс/маппинг (FR-012a).

**Alternatives considered**:
- Оставить stub — отклонено clarify.
- Отдельный raw `http.Get` в CLI — нарушает G3/G4.

---

## 9. Тестирование и секреты

**Decision**: Unit-тесты на `httptest`; токены только `test-token-…` / `fake-…`. Проверка одного исходящего `Authorization`. Для non-2xx — ровно один запрос к моку. Реальный API не обязателен в CI.

**Rationale**: constitution VII/IX; acceptance F04.

---

## 10. Итог

Открытых NEEDS CLARIFICATION нет. Готово к Phase 1 design artifacts и `/speckit-tasks`.
