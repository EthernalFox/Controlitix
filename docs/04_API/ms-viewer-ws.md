# ms-viewer WebSocket API

Назначение: realtime-доставка значений тегов из `ms-viewer` в `viewer-ui`.

## Endpoint

- URL: `/api/ws`
- Transport: WebSocket text frames (JSON)
- Все timestamps: UTC, RFC3339/RFC3339Nano

## Авторизация

- Токен передаётся в query: `?access_token=<jwt>`.
- Для fallback поддержан `Authorization: Bearer <jwt>`.
- Невалидный токен: `401` до upgrade.
- Отсутствующий токен: upgrade + close `1008` (`auth-required`).

## Протокол сообщений

Поле `t` обязательно во всех сообщениях.

### Client -> Server

```json
{ "t": "subscribe", "topics": ["tag:UUID-1", "diagram:UUID-2", "object:UUID-3"] }
{ "t": "unsubscribe", "topics": ["tag:UUID-1", "diagram:UUID-2"] }
{ "t": "ping" }
```

Поддерживаемые topic:
- `tag:<uuid>`
- `diagram:<uuid>`
- `object:<uuid>`
- `alarms`

`diagram:<uuid>` и `object:<uuid>` разворачиваются сервером во внутренний набор `tag_id`.

### Server -> Client

```json
{
  "t": "welcome",
  "session_id": "...",
  "server_time": "2026-04-28T10:00:00Z",
  "limits": { "max_subscriptions": 200, "debounce_ms": 100 }
}

{ "t": "subscribed", "topics": ["diagram:UUID-2"] }
{ "t": "unsubscribed", "topics": ["object:UUID-3"] }

{ "t": "value", "tag_id": "UUID-1", "ts": "2026-04-28T10:00:00.123Z", "v": 75.4, "q": "ok" }

{
  "t": "snapshot",
  "tag_id": "UUID-1",
  "ts": "2026-04-28T10:00:00Z",
  "v": 75.4,
  "q": "ok",
  "reason": "subscribe"
}

{
  "t": "snapshot",
  "tag_id": "UUID-1",
  "ts": "2026-04-28T10:00:00Z",
  "v": 75.4,
  "q": "ok",
  "reason": "heartbeat"
}

{
  "t": "topics_changed",
  "topic": "diagram:UUID-2",
  "added_count": 1,
  "removed_count": 0
}

{ "t": "pong" }

{
  "t": "config.changed",
  "entity_type": "diagram",
  "entity_id": "UUID-2",
  "operation": "published",
  "timestamp": "2026-04-28T10:00:00Z",
  "payload": { "diagram_id": "UUID-2" }
}

{
  "t": "error",
  "code": "limit-exceeded",
  "detail": "subscription limit exceeded",
  "topics": ["object:UUID-3"]
}
```

`error.code`:
- `limit-exceeded`
- `invalid-topic`
- `auth-required`
- `unauthorized`
- `rate-limited`

## Membership resolution

- `tag:<uuid>`: direct.
- `diagram:<uuid>`:
  - `SELECT DISTINCT tag_id FROM public.figures WHERE diagram_id=$1 AND tag_id IS NOT NULL AND deleted_at IS NULL`
- `object:<uuid>`:
  - `SELECT t.id FROM tags.tags t JOIN devices.devices d ON d.id=t.device_id WHERE d.object_id=$1 AND t.deleted_at IS NULL AND d.deleted_at IS NULL`

Ограничения:
- `MS_VIEWER_WS_GROUP_RESOLVE_LIMIT` (по умолчанию `500`) ограничивает размер развёртки одного `object:/diagram:` topic.
- При превышении лимита сервер возвращает `error.code = "limit-exceeded"` для этого topic и не подписывает его.

## Invalidation по `config.changed`

- `entity_type=diagram` -> invalidate `diagram:<entity_id>`.
- `entity_type=figure` -> invalidate `diagram:<payload.diagram_id>`.
- `entity_type=object` -> invalidate `object:<entity_id>`.
- `entity_type=device` -> invalidate `object:<payload.object_id>`.
- `entity_type=tag` -> invalidate all cached group-topics, где фигурирует `tag_id`.

После invalidation сервер пересчитывает membership активных групповых подписок и отправляет `topics_changed` с `added_count`/`removed_count`. Событие best-effort: при переполнении outbox может быть вытеснено, как и `value`.

## Поведение

- На `subscribe`: сервер отправляет `subscribed`, затем `snapshot` для новых активированных тегов (если есть last-value в Redis или in-memory cache).
- Поток `value` идёт с дебаунсом `WSDebounceMs` per `(connection, tag)`.
- Если новых значений по тегу нет 30 секунд, отправляется `snapshot` с `reason="heartbeat"`.
- `config.changed` отправляется всем активным соединениям (глобальный broadcast, без topic-подписок).
- Сервер выполняет ping каждые `WSPingIntervalSec`; при отсутствии pong в `WSPongTimeoutSec` соединение закрывается.

## Лимиты

- `WSMaxSubscriptionsPerConn` (считает уникальные `tag_id` для одного соединения)
- `WSMaxConnectionsPerUser`
- `WSWriteBufferSize` (outbox глубина per connection)
- `WSDebounceMs`
- `WSPingIntervalSec`
- `WSPongTimeoutSec`
- `WSGroupResolveLimit`

## Close codes

- `1000` normal closure
- `1001` going away (shutdown / pong timeout / connection replacement)
- `1008` policy violation (`auth-required`)
- `1011` internal error (panic/recover)

