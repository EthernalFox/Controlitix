# ms-auth API

Назначение: аутентификация пользователей, выдача JWT-токенов, управление ТУЗ и админ CRUD пользователей/ролей.

Базовый префикс через gateway: `/api/auth`.
Базовый префикс для JWKS: `/.well-known/jwks.json` (пробрасывается nginx без `/api/auth` — чтобы соответствовать стандартам OAuth2/OIDC).

Архитектура: [`../02_Архитектура/Микросервисы/ms-auth.md`](../02_Архитектура/Микросервисы/ms-auth.md). Решения: [ADR-0007](../08_ADR/ADR-0007-service-accounts-env-secrets.md), [ADR-0012](../08_ADR/ADR-0012-ms-auth-pluggable-identity.md).

## Общие соглашения

- `Content-Type: application/json` для всех запросов и ответов, кроме JWKS.
- Ошибки — RFC 7807 (`application/problem+json`), как в ms-editor.
- Access-токен передаётся в заголовке `Authorization: Bearer <jwt>`.
- Refresh-токен — opaque-строка (base64url, 43 символа). Возвращается в теле ответа. Клиент (editor-ui / viewer-ui) хранит его в httpOnly+Secure+SameSite=Strict cookie, которую nginx проставляет при ответе `/login` и `/refresh`. Серверная реализация принимает refresh как из тела, так и из cookie — для гибкости.

## 1. Пользовательская аутентификация

### 1.1 `POST /api/auth/login`

Password grant. В MVP — единственный способ логина.

Запрос:

```json
{
  "username": "admin",
  "password": "••••••••"
}
```

Ответ `200 OK`:

```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "fZkU...43chars",
  "user": {
    "id": "9a4f...",
    "subject": "local:9a4f...",
    "username": "admin",
    "display_name": "Администратор",
    "email": "admin@example.local",
    "roles": ["admin"],
    "source": "local"
  }
}
```

Ошибки:

| Код | `type`                                       | Случай |
|-----|----------------------------------------------|--------|
| 400 | `/errors/validation`                         | Пустой/некорректный payload |
| 401 | `/errors/auth/invalid-credentials`           | Неверная пара логин/пароль (единый ответ для unknown user и wrong password). |
| 403 | `/errors/auth/user-disabled`                 | `users.is_active = false`. |
| 429 | `/errors/auth/rate-limited`                  | Слишком много попыток (заголовок `Retry-After`). |

### 1.2 `POST /api/auth/refresh`

Ротация refresh + выдача нового access. Повторное использование уже использованного refresh → инвалидация всей family_id.

Запрос (любой из вариантов):

```json
{ "refresh_token": "fZkU..." }
```

Либо — refresh из cookie `refresh_token`.

Ответ `200 OK`:

```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "NEW43chars..."
}
```

Ошибки:

| Код | `type` |
|-----|--------|
| 400 | `/errors/validation` |
| 401 | `/errors/auth/invalid-refresh` — не найден, истёк, отозван, переиспользован |

### 1.3 `POST /api/auth/logout`

Отзыв текущей family_id.

Запрос (тело опционально, если cookie):

```json
{ "refresh_token": "fZkU..." }
```

Ответ `204 No Content`. Всегда 204, даже если токен неизвестен (не даём сигнала атакующему).

### 1.4 `GET /api/auth/userinfo`

Текущий пользователь из access-токена.

Заголовок: `Authorization: Bearer <access>`.

Ответ `200 OK`:

```json
{
  "id": "9a4f...",
  "subject": "local:9a4f...",
  "username": "admin",
  "display_name": "Администратор",
  "email": "admin@example.local",
  "roles": ["admin"],
  "source": "local",
  "last_login_at": "2026-04-20T09:41:00Z"
}
```

Ошибки:

| Код | `type` |
|-----|--------|
| 401 | `/errors/auth/invalid-token` |
| 401 | `/errors/auth/token-expired` |

## 2. Межсервисная аутентификация (ТУЗ)

### 2.1 `POST /api/auth/service-token`

Client credentials grant. Refresh не выдаётся.

Запрос:

```json
{
  "grant_type": "client_credentials",
  "client_id": "ms-poll",
  "client_secret": "••••••••",
  "scope": "tags.values.write config.read"
}
```

Поле `scope` опционально. Если указано — запрашиваемые scope должны быть подмножеством зарегистрированных за `client_id`; иначе 400 `invalid_scope`.

Ответ `200 OK`:

```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 900,
  "scope": "tags.values.write config.read"
}
```

Ошибки:

| Код | `type` |
|-----|--------|
| 400 | `/errors/auth/unsupported-grant` — `grant_type` не `client_credentials` |
| 400 | `/errors/auth/invalid-scope` — запрошен неразрешённый scope |
| 401 | `/errors/auth/invalid-client` |
| 403 | `/errors/auth/client-disabled` |

## 3. JWKS

### 3.1 `GET /.well-known/jwks.json`

Стандартный JWKS-документ для валидации JWT. Проксируется nginx напрямую на ms-auth, **без префикса `/api/auth`** — соответствует [RFC 8414](https://datatracker.ietf.org/doc/html/rfc8414).

Ответ `200 OK`, `Content-Type: application/json`:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "2026-04-20",
      "n": "0vx7agoebGcQSuuPiLJXZpt...",
      "e": "AQAB"
    }
  ]
}
```

Может содержать несколько ключей во время ротации. Кэшировать с учётом заголовков `Cache-Control`.

## 4. Admin API

Все эндпоинты требуют JWT с ролью `admin`. Префикс `/api/auth/admin`.

### 4.1 Пользователи

- `GET    /api/auth/admin/users?offset=&limit=&q=&source=&is_active=` — список с фильтрами.
- `POST   /api/auth/admin/users` — создать локального пользователя.
- `GET    /api/auth/admin/users/{id}` — получить.
- `PATCH  /api/auth/admin/users/{id}` — обновить `display_name`, `email`, `is_active`, сменить пароль.
- `DELETE /api/auth/admin/users/{id}` — soft delete (`is_active = false` + revoke всех refresh).
- `POST   /api/auth/admin/users/{id}/roles` — назначить роли.
- `DELETE /api/auth/admin/users/{id}/roles/{role}` — снять роль.
- `POST   /api/auth/admin/users/{id}/revoke-sessions` — отозвать все refresh пользователя.

Запрос `POST /admin/users`:

```json
{
  "username": "ivan.petrov",
  "display_name": "Иван Петров",
  "email": "ivan.petrov@example.local",
  "password": "••••••••",
  "roles": ["engineer"]
}
```

Ответ `201 Created`: объект User (без `password_hash`).

### 4.2 Роли

В MVP роли — фиксированное множество (`admin`, `engineer`, `operator`), поэтому endpoint только для чтения:

- `GET /api/auth/admin/roles` — список ролей с описаниями.

Создание новых ролей — вне MVP.

### 4.3 Источники идентичности (будущее)

Зарезервированные эндпоинты. В MVP возвращают единственный источник `local`. CRUD для LDAP/AD — вне MVP.

- `GET  /api/auth/admin/identity-sources` — список.
- `POST /api/auth/admin/identity-sources` — **501 Not Implemented** в MVP.
- `GET  /api/auth/admin/identity-sources/{id}`.

### 4.4 Маппинг ролей (будущее)

Заготовка под AD/LDAP-группы:

- `GET    /api/auth/admin/identity-sources/{id}/role-mappings` — пустой массив в MVP.
- `POST   /api/auth/admin/identity-sources/{id}/role-mappings` — **501** в MVP.
- `DELETE /api/auth/admin/identity-sources/{id}/role-mappings/{mapping_id}` — **501** в MVP.

### 4.5 ТУЗ (service accounts)

- `GET    /api/auth/admin/service-accounts`
- `POST   /api/auth/admin/service-accounts` — при создании возвращает `client_secret` **один раз** (plaintext). Далее хранится только хэш.
- `GET    /api/auth/admin/service-accounts/{id}`
- `PATCH  /api/auth/admin/service-accounts/{id}` — включить/выключить, обновить scope.
- `POST   /api/auth/admin/service-accounts/{id}/rotate-secret` — возвращает новый `client_secret` один раз.
- `DELETE /api/auth/admin/service-accounts/{id}`

### 4.6 Аудит

- `GET /api/auth/admin/audit?offset=&limit=&action=&actor=&from=&to=` — чтение `auth.audit_log`.

## 5. Health

- `GET /healthz` — 200 `{"status":"ok"}`.
- `GET /readyz` — 200 если доступны PG и Kafka producer; 503 иначе.

## 6. Будущие эндпоинты (не в MVP)

Фиксируются здесь, чтобы клиенты (`editor-ui`, `viewer-ui`) заранее учитывали совместимый дизайн:

- `GET  /api/auth/login/options` — список доступных способов логина (в MVP — `{"methods":["password"]}`, позже — `password`, `spnego`, `oidc:<provider>`).
- `POST /api/auth/login/spnego` — Kerberos/SPNEGO (`Authorization: Negotiate <ticket>`).
- `GET  /api/auth/oidc/{provider}/start` — инициация authorization_code+PKCE.
- `GET  /api/auth/oidc/{provider}/callback` — OIDC callback.

Эти эндпоинты в MVP возвращают `404` или `501`, клиенты не должны на них опираться.

## 7. Формат JWT (для приёмников)

Access-токен:

```json
{
  "iss": "controlitix-auth",
  "aud": ["controlitix-api"],
  "sub": "local:9a4f...",
  "exp": 1745138460,
  "iat": 1745137560,
  "jti": "...",
  "src": "local",
  "roles": ["admin"],
  "scope": "",
  "username": "admin",
  "display_name": "Администратор"
}
```

Service-токен:

```json
{
  "iss": "controlitix-auth",
  "aud": ["controlitix-internal"],
  "sub": "svc:ms-poll",
  "exp": 1745138460,
  "iat": 1745137560,
  "jti": "...",
  "src": "local",
  "roles": [],
  "scope": "tags.values.write config.read",
  "client_id": "ms-poll"
}
```

Приёмники валидируют: подпись через JWKS, `iss`, `aud`, `exp`, `nbf`/`iat`. `sub` не интерпретируется кроме как для логирования.

## Связанные документы

- [`../02_Архитектура/Микросервисы/ms-auth.md`](../02_Архитектура/Микросервисы/ms-auth.md)
- [`../02_Архитектура/Безопасность.md`](../02_Архитектура/Безопасность.md)
- [ADR-0007](../08_ADR/ADR-0007-service-accounts-env-secrets.md)
- [ADR-0012](../08_ADR/ADR-0012-ms-auth-pluggable-identity.md)
