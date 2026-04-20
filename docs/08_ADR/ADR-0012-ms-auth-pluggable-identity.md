# ADR-0012: ms-auth — JWT + расширяемые источники идентичности

- Статус: accepted
- Дата: 2026-04-20
- Авторы: команда Controlitix

## Контекст

[ADR-0007](ADR-0007-service-accounts-env-secrets.md) уже зафиксировал, что межсервисная и пользовательская аутентификация строятся на **JWT + JWKS** и выдаёт их `ms-auth`. Этот ADR детализирует устройство самого `ms-auth` и отвечает на вопрос, **как не переписывать сервис** при последующем подключении AD / LDAP / Kerberos / внешнего OIDC-провайдера.

В MVP явно зафиксировано отсутствие SSO/LDAP/AD ([`../01_Цель-и-область/НФТ.md`](../01_Цель-и-область/НФТ.md) — раздел «Ограничения MVP»). Однако пилотные и тем более крупные заказчики почти гарантированно принесут корпоративный каталог. Если заложить соответствующую абстракцию с первой строки кода — это малая инкрементальная стоимость; если нет — интеграция превратится в переписывание таблицы `users`, миграцию уже выпущенных токенов и ломку refresh-потока.

## Рассмотренные варианты

### Модель хранения идентичности

1. **Плоская модель `users(email, password_hash, role)`**: только локальные пользователи, `user_id = BIGSERIAL`. Просто, минимум кода. Проблема — `sub` в JWT привязан к локальному PK; добавление AD/LDAP требует либо дублирования пользователей (и конфликтов), либо миграции всех существующих `sub`.
2. **Модель с внешним `subject` и справочником источников**: `users(subject TEXT UNIQUE, source_id, password_hash NULL, ...)` + `identity_sources(type, config)` + `role_mappings`. Локальные пользователи — частный случай. LDAP/AD/Kerberos — просто новый `type` источника.
3. **Вынос идентичности целиком за пределы `ms-auth`** (Keycloak / Authentik / Dex как IdP, `ms-auth` только валидирует). Правильно в долгой перспективе, избыточно на MVP с 2–3 ролями и отсутствием внешнего IdP у первого заказчика.

### Где хэшировать / хранить пароли

1. **bcrypt (стандартный выбор)** — проверенный, но медленный и ограничен 72 байтами.
2. **Argon2id (OWASP-рекомендация)** — современный, параметры настраиваются (m/t/p), переносим между версиями, устойчив к GPU.
3. **Vault / KMS для паролей** — **не подходит**: пароли не восстанавливаемы по определению, Vault — это хранилище **секретов**, которые нужно прочитать. Хэш паролей идёт в БД, а pepper — рядом с приватным ключом JWT.

### Кто валидирует JWT

1. **Gateway (Nginx с модулем JWT)** — централизованно, но коммерческая опция Nginx Plus или стороннее lua/njs. ADR-0006 явно запрещает валидацию на gateway.
2. **Каждый сервис сам** — через общий Go-пакет (`shared/authctx`) с кэшем JWKS. Соответствует ADR-0006.

## Решение

### Архитектура `ms-auth`

```
ms-auth/
  cmd/ms-auth/main.go
  internal/
    domain/                 # User, Role, ServiceAccount, Token, ExternalIdentity
    usecase/
      auth_service.go       # оркестрация: login → provider → token issuance
      token_service.go      # JWT-выпуск, refresh rotation, reuse detection
      user_service.go       # admin CRUD пользователей
    infrastructure/
      repository/           # только SELECT/INSERT/UPDATE таблиц auth.*
      identity/             # реализации IdentityProvider
        local.go            # LocalProvider (Argon2id)
        ldap.go             # future — интерфейс есть, реализации нет
        ad.go               # future
        kerberos.go         # future
        oidc.go             # future (federation)
      jwks/                 # RS256 keystore + /.well-known/jwks.json
    api/http/
      auth.go               # /login, /refresh, /logout, /userinfo
      service.go            # /service-token (client_credentials)
      admin.go              # /admin/users/*, /admin/roles/*
      health.go             # /healthz, /readyz
```

### Токены

- **Алгоритм:** RS256 (2048-bit). Приватный ключ — PEM в env `MS_AUTH_JWT_PRIVATE_KEY` (или путь `_PATH`), публичный ключ генерируется из приватного и отдаётся через `GET /.well-known/jwks.json` с `kid`.
- **Claims:**
  - `iss` — `controlitix-auth`.
  - `aud` — массив. Для user-JWT: `["controlitix-api"]`. Для service-JWT: `["controlitix-internal"]`.
  - `sub` — **стабильный строковый идентификатор**: `local:<uuid>`, `ldap:<dn-hash>`, `ad:<objectGUID>`, `krb:<principal>`, `svc:<service-name>`. **Не BIGSERIAL**.
  - `exp`, `iat`, `jti`.
  - `scope` (строка, OAuth2-стиль) и `roles` (массив строк: `admin|engineer|operator`).
  - `src` — тип источника идентичности (`local|ldap|ad|kerberos|oidc`), нужен для аудита.
- **TTL:** access — 15 мин, refresh — 14 дней.
- **Ротация ключей JWKS:** несколько `kid` в JWKS; приёмники держат ключи до `exp` самого долгоживущего refresh + запас.

### Refresh-токены

- Хранятся в БД (`auth.refresh_tokens`), НЕ в JWT-формате. Отдаются клиенту как opaque-строка (32 байта random → base64url).
- **Ротация при каждом использовании** (refresh-rotation): выдача нового access **и нового refresh**, старый помечается `used_at`.
- **Reuse detection:** повторное предъявление уже использованного refresh — инвалидация всей «семьи» (`family_id`), форс-логаут пользователя. Стандартный ответ на кражу токена.

### Модель идентичности

Интерфейс внутри `ms-auth`:

```go
type Credentials struct {
    Username string       // логин/email/UPN
    Password string       // для password-flow
    Extra    map[string]any // для SPNEGO-тикета, OIDC-кода и т.д.
}

type ExternalIdentity struct {
    Subject     string        // уникально в рамках источника
    DisplayName string
    Email       string
    Groups      []string      // внешние группы/роли
    Raw         map[string]any
}

type IdentityProvider interface {
    Type() string                       // "local" | "ldap" | "ad" | ...
    Authenticate(ctx, Credentials) (*ExternalIdentity, error)
}
```

**В MVP реализован только `LocalProvider`.** Остальные реализации — заглушки/отсутствуют, но интерфейс и точки подключения существуют, включая таблицу `auth.identity_sources` (в MVP — одна строка `type=local`) и `auth.role_mappings` (маппинг внешних групп в наши роли, в MVP пустая).

После успешного `Authenticate` применяется общий код:

1. `Lookup or JIT-create` пользователя в `auth.users` по `(source_id, subject)`.
2. Разрешение ролей: локальные роли + маппинг `ExternalIdentity.Groups` через `role_mappings`.
3. Выдача access+refresh токенов.
4. Запись события в `auth.audit_log` и Kafka `audit.logs`.

### Хэширование паролей

- **Algorithm:** Argon2id.
- **Параметры (MVP):** `m=64 MiB`, `t=3`, `p=2`, длина соли 16 байт, длина хэша 32 байта. Выбор основан на OWASP Password Storage Cheat Sheet; занимает ~100 мс на современном CPU.
- **Pepper:** глобальная секретная строка из env `MS_AUTH_PASSWORD_PEPPER` (≥ 32 случайных байта base64). Хранится отдельно от БД — компрометация одной БД без env не даёт провести offline-брутфорс.
- **Формат хранения:** стандартный Argon2 encoded `$argon2id$v=19$m=...,t=...,p=...$salt$hash`, позволяет сменить параметры без слома существующих хэшей.

### Межсервисная аутентификация (ТУЗ)

Из ADR-0007. В `ms-auth` это отдельный код-путь:

- `POST /api/auth/service-token` принимает `grant_type=client_credentials`, `client_id`, `client_secret`.
- `client_id` и хэш `client_secret` — в таблице `auth.service_accounts`.
- Выдаётся access-JWT с `sub=svc:<name>`, `aud=["controlitix-internal"]`, `scope` из записи в БД.
- Refresh для ТУЗ **не выдаётся** — сервис обновляет access по мере истечения, перепредъявляя client_credentials.

### Валидация на стороне приёмника

Общий Go-пакет `shared/authctx` (живёт в моно-репо, подключается каждым сервисом):

- Кэш JWKS с TTL 15 мин + фоновый refresh + fallback на stale-cache при недоступности ms-auth.
- Middleware `RequireAuth(audience, scopes...)` для chi.
- Экстрактор `FromContext(ctx) → Principal{Subject, Roles, Scopes, Source}`.
- **Не валидирует revocation** — access-token короткоживущий. Отзыв реализуется через отзыв refresh (удаление из БД); access просто истекает за 15 минут.

## Последствия

**Позитивно:**

- Подключение LDAP/AD/Kerberos/OIDC = **новый файл в `internal/infrastructure/identity/` и строка в `identity_sources`**, без изменений токенов, БД, API.
- Стабильный `sub` → токены не инвалидируются при смене источника (например, при миграции пользователя из `local` в `ldap`).
- Refresh-rotation + reuse detection = стандартная защита от кражи токена без сложной инфраструктуры.
- Pepper отдельно от БД = один вектор компрометации не даёт offline-brute-force.
- Единый контракт JWT для user и service — валидатор один.

**Негативно:**

- На MVP это «недоиспользованная» архитектура: одна таблица `identity_sources` с единственной строкой, пустая `role_mappings`, интерфейс с единственной реализацией. Стоимость — примерно +1 день к реализации.
- JIT-provisioning при подключении внешних источников потребует аудита: что делать, если пользователь со временем меняет группы / удаляется в AD. Это решается отдельным ADR в момент подключения.
- Ротация ключей JWKS — ручная в MVP (замена `MS_AUTH_JWT_PRIVATE_KEY` + перезапуск + `kid` bump). Автоматическая ротация — вне MVP.

**Исключено из MVP (остаётся в `internal/infrastructure/identity/` как отсутствующие реализации):**

- LDAP-провайдер (bind + search group membership).
- Active Directory (LDAP + objectGUID + nested groups).
- Kerberos/SPNEGO (GSSAPI + keytab).
- OIDC federation (authorization_code + PKCE к внешнему IdP).
- MFA / TOTP.
- Per-object RBAC (только глобальные роли admin/engineer/operator).
- Password reset через email.
- Account lockout после N неудач (будет — rate-limit на `/login`, без локаута).

## Миграционный путь LDAP/AD/Kerberos (для будущего)

Фиксируется как ориентир. При переходе на следующий релиз будет отдельный ADR с деталями.

1. **LDAP/AD:** реализовать `LDAPProvider.Authenticate`: bind по `Username`, search группы по `memberOf`. Добавить строку в `identity_sources` с конфигом (bind DN, base DN, search filter, TLS-параметры). Заполнить `role_mappings`. Пользователь при первом логине создаётся JIT в `auth.users` с `subject=ad:<objectGUID>` и `password_hash=NULL`.
2. **Kerberos/SPNEGO:** ручка `POST /api/auth/login/spnego` принимает `Authorization: Negotiate <base64-ticket>`, валидирует по keytab (`gokrb5`), эмитит обычный JWT. Браузер настраивается корпоративной политикой. Остальные сервисы об этом не знают — валидация JWT не меняется.
3. **OIDC federation (Keycloak/ADFS/AzureAD):** ручки `/oidc/start` и `/oidc/callback` с `authorization_code + PKCE`. Внешний IdP — в `identity_sources` с его `discovery_url`. Тот же JIT.
4. **Миграция локальных пользователей** в AD — разрешена: сохраняется `auth.users`, меняется только `source_id` и `subject`. Локальный `password_hash` занулится. Существующие refresh-токены инвалидируются одним SQL (`UPDATE ... SET revoked_at = now()`).

## Связанные документы

- [ADR-0006](ADR-0006-nginx-gateway.md) — Nginx не валидирует JWT.
- [ADR-0007](ADR-0007-service-accounts-env-secrets.md) — ТУЗ и секреты в env.
- [`../02_Архитектура/Микросервисы/ms-auth.md`](../02_Архитектура/Микросервисы/ms-auth.md) — детальная архитектура сервиса.
- [`../02_Архитектура/Безопасность.md`](../02_Архитектура/Безопасность.md) — модель угроз и контроли.
- [`../04_API/ms-auth.md`](../04_API/ms-auth.md) — HTTP-контракт.
- [`../01_Цель-и-область/НФТ.md`](../01_Цель-и-область/НФТ.md) — раздел «Безопасность (MVP)».
