# ms-auth — аутентификация и авторизация

## Назначение

- Единая точка аутентификации пользователей и сервисов Controlitix.
- Выпуск JWT (user + service), их валидация через JWKS, ротация refresh-токенов.
- Управление пользователями, ролями, технологическими учётными записями (ТУЗ).
- Аудит событий безопасности.

## Границы ответственности

- **Что делает:** выпускает/валидирует токены, хранит пользователей и ТУЗ, проверяет пароли, управляет ролями, пишет аудит.
- **Что не делает:** не валидирует per-object RBAC (в MVP ролей три и они глобальные), не хранит профили пользователей других сервисов, не является policy decision point для прикладных правил (каждый сервис решает сам на основе `roles`/`scope`), не занимается session management в понимании веб-cookie (клиент держит refresh-token).

## Интерфейсы

### HTTP (см. [`../../04_API/ms-auth.md`](../../04_API/ms-auth.md))

- `POST /api/auth/login` — password grant → access + refresh.
- `POST /api/auth/refresh` — обмен refresh на новую пару (с ротацией).
- `POST /api/auth/logout` — отзыв refresh-семьи.
- `GET  /api/auth/userinfo` — текущий пользователь по access-токену.
- `POST /api/auth/service-token` — client_credentials для ТУЗ.
- `GET  /.well-known/jwks.json` — публичные ключи для валидации.
- `/api/auth/admin/*` — CRUD пользователей, ролей, ТУЗ (только `admin`).
- `/healthz`, `/readyz` — health-пробы.

### Kafka

- **Publishes:** `audit.logs` — все события аутентификации, выдачи токенов, админ-действия (см. [ADR-0010](../../08_ADR/ADR-0010-kafka-event-bus.md)).
- **Consumes:** ничего.

### База данных

- Схема `auth` (эксклюзивно принадлежит `ms-auth`, ни один другой сервис туда не пишет).
- Миграции через Goose в `ms-auth/migrations/`.

## Ключевые архитектурные решения

Детально — [ADR-0012](../../08_ADR/ADR-0012-ms-auth-pluggable-identity.md).

1. **JWT RS256 + JWKS.** Валидация на стороне приёмников (ADR-0006), не на gateway.
2. **Pluggable identity sources.** Интерфейс `IdentityProvider`; в MVP реализован только `LocalProvider`, но таблицы и точки расширения заложены под LDAP / AD / Kerberos / OIDC.
3. **Стабильный `subject`.** Внешний идентификатор вида `local:<uuid>` / `ldap:<dn-hash>` / `ad:<objectGUID>` / `krb:<principal>` / `svc:<name>` — не PK БД. Это якорь для будущей миграции пользователей между источниками без ломки токенов.
4. **Refresh-rotation + reuse detection.** Каждое использование refresh → новый refresh; повтор старого → инвалидация всей семьи.
5. **Argon2id + pepper.** Пароли хэшируются Argon2id; глобальный pepper хранится в env отдельно от БД.
6. **ТУЗ через client_credentials.** Межсервисные токены — тот же JWT, другой `aud`, без refresh.

## Доменная модель (high-level)

```
User
  id: uuid
  subject: text  ("local:<uuid>", "ldap:<hash>", ...)
  source_id: int → IdentitySource
  username, email, display_name
  password_hash: text | null   (null для внешних источников)
  is_active: bool
  created_at, updated_at, last_login_at
  ──< UserRole >── Role (admin | engineer | operator)

IdentitySource
  id: int, type: text (local|ldap|ad|kerberos|oidc)
  config: jsonb       (bind DN, base DN, realm, и т.п.)
  is_enabled: bool

RoleMapping
  source_id → IdentitySource
  external_group: text      (CN AD-группы и т.п.)
  role: text                (наша роль)

ServiceAccount
  id: uuid
  client_id: text unique
  client_secret_hash: text   (Argon2id)
  scopes: text[]
  is_active: bool

RefreshToken
  id: uuid
  user_id → User
  family_id: uuid            (общий для всех токенов в цепочке ротации)
  token_hash: text           (SHA-256 от opaque-строки)
  issued_at, expires_at
  used_at: timestamptz | null
  revoked_at: timestamptz | null
  user_agent, ip: для аудита

AuditEvent
  id: uuid
  actor_subject: text
  action: text  (login.success, login.failure, token.refresh, admin.user.create, ...)
  target: text | null
  result: text  (success|failure)
  reason: text | null
  metadata: jsonb
  occurred_at: timestamptz
```

## Поток: пользовательский логин

```
editor-ui/viewer-ui → POST /api/auth/login {username, password}
    └─ AuthService.Login
        ├─ Выбирает IdentitySource (в MVP всегда local)
        ├─ Provider.Authenticate(creds) → ExternalIdentity
        │    (LocalProvider: Argon2id.Verify(password, user.password_hash))
        ├─ Lookup or JIT-create в auth.users
        ├─ Разрешение ролей (локальные + маппинг групп)
        ├─ TokenService.IssueAccess(user, roles, scope)     → RS256-JWT
        ├─ TokenService.IssueRefresh(user, family_id=new)   → opaque, сохранён в auth.refresh_tokens
        └─ audit.login.success / failure → PG + Kafka audit.logs
```

Rate limit: N ошибок логина за M минут → 429 (Redis-счётчик по связке IP + username).

## Поток: обновление токена

```
client → POST /api/auth/refresh {refresh_token}
    └─ TokenService.Rotate
        ├─ Находит запись по hash(refresh_token).
        ├─ Если used_at != null → reuse detected → revoke вся family_id → 401.
        ├─ Если revoked_at != null или истёк → 401.
        ├─ Помечает текущий used_at = now().
        ├─ Выпускает новую пару (access + новый refresh), family_id тот же.
        └─ audit.token.refresh.
```

## Поток: ТУЗ

```
ms-poll (на старте) → POST /api/auth/service-token
    grant_type=client_credentials
    client_id=ms-poll, client_secret=<env>
  └─ Проверка auth.service_accounts (Argon2id.Verify client_secret).
  └─ Access-JWT (aud=controlitix-internal, sub=svc:ms-poll, scope=...).
  └─ Refresh не выдаётся. Сервис перепредъявляет client_credentials при истечении.
```

## Поток: интеграция LDAP/AD (будущее, не в MVP)

```
editor-ui → POST /api/auth/login {username, password}
  └─ AuthService резолвит IdentitySource по username (например, UPN → source=ad)
      └─ LDAPProvider.Authenticate:
           - bind по (DN пользователя, password)
           - search memberOf → groups
           - возврат ExternalIdentity{subject=ad:<objectGUID>, groups}
      └─ JIT-create or lookup в auth.users
      └─ Роли = RoleMapping.resolve(ad-groups) ∪ локальные user_roles
      └─ Выпуск токенов, тот же код-путь, что и local.
```

Ни один другой сервис не меняется.

## Поток: Kerberos/SPNEGO (будущее, не в MVP)

```
Browser → GET <editor-ui>  (через корпоративный Windows-домен)
    └─ Nginx проксирует "Authorization: Negotiate <ticket>" на ms-auth.
ms-auth → POST /api/auth/login/spnego
    └─ KerberosProvider.Authenticate через gokrb5 + keytab
    └─ ExternalIdentity{subject=krb:user@REALM}
    └─ тот же код-путь, выдача JWT.
```

## Зависимости

- **PostgreSQL** — схема `auth`.
- **Kafka** — топик `audit.logs`.
- **Redis** — rate-limit счётчики (опционально в MVP; если нет — использовать in-memory с потерей при рестарте, допустимо на одном хосте).

## Нефункциональные требования

| Метрика | Цель MVP |
|---|---|
| Латентность `/login` | ≤ 200 мс (p95), из них ~100 мс — Argon2id |
| Латентность `/refresh` | ≤ 50 мс (p95) |
| Латентность валидации JWT приёмником | ≤ 1 мс (p99) — локальная проверка подписи |
| TTL access-токена | 15 мин |
| TTL refresh-токена | 14 дней |
| Время распространения отзыва refresh | мгновенно (БД); access живёт до истечения |
| Максимальное число активных refresh-семей на пользователя | не ограничено в MVP, мониторим |

## Ограничения MVP

Явно не реализуется:

- SSO / LDAP / AD / Kerberos / OIDC federation — **интерфейс и таблицы заложены, реализаций нет**.
- MFA / TOTP / WebAuthn.
- Self-service password reset.
- Per-object RBAC.
- Автоматическая ротация JWKS-ключей.
- Account lockout.
- Device fingerprinting.

## Связанные документы

- [ADR-0007](../../08_ADR/ADR-0007-service-accounts-env-secrets.md) — ТУЗ и секреты.
- [ADR-0012](../../08_ADR/ADR-0012-ms-auth-pluggable-identity.md) — архитектурное решение по `ms-auth`.
- [`../Безопасность.md`](../Безопасность.md) — модель угроз и контроли.
- [`../../04_API/ms-auth.md`](../../04_API/ms-auth.md) — HTTP-контракт.
- [`../../05_Сценарии-использования/Администратор_управление_доступом_RBAC.md`](../../05_Сценарии-использования/Администратор_управление_доступом_RBAC.md).
