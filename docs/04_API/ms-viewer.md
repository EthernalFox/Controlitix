# ms-viewer API

Назначение: чтение трендов, поиск тегов, realtime WebSocket и read-only доступ к опубликованным мнемосхемам.

Базовый префикс через gateway: `/api`.

## Общие соглашения

- Все `/api/*` требуют `Authorization: Bearer <jwt>`.
- JWT валидируется через `shared/authctx` и JWKS (`MS_VIEWER_JWKS_URL`).
- Ошибки возвращаются в формате RFC 7807 (`application/problem+json`).

## 1. `GET /api/trends/{tagId}`

Параметры:

- `from` (RFC3339, обязателен)
- `to` (RFC3339, обязателен)
- `step` (duration, опционально: `5s`, `1m`, `1h`)
- `agg` (`last|avg|min|max`, по умолчанию `avg`)
- `limit` (по умолчанию `1000`, максимум `5000`)

Ответ `200 OK`:

```json
{
  "tag_id": "uuid",
  "tag_name": "boiler_1.t_out",
  "device_id": "uuid",
  "device_name": "Boiler-1",
  "unit": { "id": 7, "name": "°C", "symbol": "°C", "category": "temperature" },
  "data_type": { "id": 2, "name": "float32" },
  "from": "2026-04-25T10:00:00Z",
  "to": "2026-04-25T11:00:00Z",
  "step": "10s",
  "agg": "avg",
  "source": "raw",
  "points": [
    { "ts": "2026-04-25T10:00:00Z", "v": 75.2, "q": "ok" },
    { "ts": "2026-04-25T10:00:10Z", "v": null, "q": "comm_loss" }
  ]
}
```

`source`:

- `raw` при окне `<= 24h`
- `agg_1m` при окне `> 24h`

## 2. `GET /api/trends`

Параметры:

- `tag_ids` (обязателен, CSV UUID, максимум 10)
- `from`, `to`, `step`, `agg`, `limit` — как у `/api/trends/{tagId}`

Ответ `200 OK`:

```json
{
  "series": [
    {
      "tag_id": "uuid",
      "tag_name": "boiler_1.t_out",
      "device_id": "uuid",
      "device_name": "Boiler-1",
      "unit": { "id": 7, "name": "°C", "symbol": "°C", "category": "temperature" },
      "data_type": { "id": 2, "name": "float32" },
      "from": "2026-04-25T10:00:00Z",
      "to": "2026-04-25T11:00:00Z",
      "step": "10s",
      "agg": "avg",
      "source": "raw",
      "points": []
    }
  ],
  "errors": [
    {
      "tag_id": "uuid",
      "problem": {
        "type": "/errors/trends/tag-not-found",
        "title": "Tag not found",
        "status": 404,
        "detail": "tag is missing"
      }
    }
  ]
}
```

Поле `errors[]` — расширение batch-контракта: ошибка одного тега не роняет весь запрос.

## 3. `GET /api/tags`

Read-only поиск для TrendTagPicker.

Параметры:

- `search` (опционально)
- `limit` (опционально, максимум 20)

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": "uuid",
      "name": "boiler_1.t_out",
      "device_id": "uuid",
      "device_name": "Boiler-1",
      "unit_symbol": "°C"
    }
  ],
  "total": 1
}
```

## 4. Diagrams (read-only)

### `GET /api/objects`

Список объектов, для которых есть хотя бы одна опубликованная мнемосхема.

Параметры:

- `limit` (по умолчанию `50`, максимум `200`)
- `offset` (по умолчанию `0`)

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": "uuid",
      "name": "Котельная №3",
      "description": "Северный корпус",
      "published_diagram_count": 2,
      "first_published_diagram_id": "uuid"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 50
}
```

### `GET /api/objects/{objectId}/diagrams`

Список опубликованных мнемосхем объекта.

Параметры:

- `limit` (по умолчанию `50`, максимум `200`)
- `offset` (по умолчанию `0`)

Ответ `200 OK`:

```json
{
  "items": [
    {
      "id": "uuid",
      "object_id": "uuid",
      "name": "Котёл №1",
      "description": "Главная схема",
      "published_at": "2026-04-25T08:00:00Z",
      "figure_count": 42,
      "bound_tag_count": 18
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 50
}
```

### `GET /api/diagrams/{diagramId}`

Полная опубликованная мнемосхема со всеми фигурами.

Заголовок ответа:

- `Cache-Control: private, max-age=10`

Ответ `200 OK`:

```json
{
  "id": "uuid",
  "object_id": "uuid",
  "object_name": "Котельная №3",
  "name": "Котёл №1",
  "description": "Главная схема",
  "published_at": "2026-04-25T08:00:00Z",
  "canvas": {
    "width": 1920,
    "height": 1080,
    "background": "#F5F5F5"
  },
  "figures": [
    {
      "id": "uuid",
      "type": "tag",
      "tag_id": "uuid",
      "params": {
        "x": 100,
        "y": 120,
        "width": 200,
        "height": 42,
        "text": "T_out",
        "fill": "#ffffff",
        "stroke": "#101010",
        "strokeWidth": 1,
        "dynamics": {
          "fillByQuality": {
            "ok": "#4CAF50",
            "hi": "#FFA726",
            "hihi": "#EF5350",
            "uncertain": "#42A5F5",
            "bad": "#78909C",
            "comm_loss": "#546E7A",
            "offline": "#37474F",
            "acknowledged": "#7E57C2"
          },
          "textFromValue": {
            "format": "%.1f",
            "appendUnit": true
          },
          "visibilityByQuality": ["ok", "hi", "hihi"],
          "blinkOnAlarm": true
        }
      },
      "tag": {
        "id": "uuid",
        "name": "boiler_1.t_out",
        "device_id": "uuid",
        "device_name": "Boiler-1",
        "unit": {
          "id": 7,
          "name": "°C",
          "symbol": "°C",
          "category": "temperature"
        },
        "data_type": {
          "id": 2,
          "name": "float32"
        }
      }
    }
  ]
}
```

`canvas` в ответе возвращается всегда. Источник: поля из схемы (если присутствуют) или fallback:

- `width=1920`
- `height=1080`
- `background="#F5F5F5"`

### `GET /api/diagrams/{diagramId}/snapshot`

Снимок последних значений всех тегов, привязанных к фигурам диаграммы.

Источник данных: только Redis (`ms-viewer:last:<tag_id>`), без fallback в PostgreSQL.

Ответ `200 OK`:

```json
{
  "diagram_id": "uuid",
  "ts": "2026-04-25T10:00:00.123Z",
  "values": [
    {
      "tag_id": "uuid",
      "ts": "2026-04-25T10:00:00.000Z",
      "v": 75.2,
      "q": "ok"
    }
  ],
  "missing_tag_ids": ["uuid"]
}
```

## 5. Health

- `GET /healthz` > `200 {"status":"ok"}`
- `GET /readyz` > `200`, если доступны PostgreSQL и Redis, иначе `503`

## Ошибки (RFC 7807)

- `400 /errors/trends/invalid-range`
- `404 /errors/trends/tag-not-found`
- `403 /errors/trends/forbidden`
- `400 /errors/diagrams/invalid-request`
- `404 /errors/diagrams/not-found`
- `404 /errors/diagrams/not-published`
- `503 /errors/realtime/cache-unavailable`
- `401 /errors/auth/invalid-token`
- `401 /errors/auth/token-expired`
- `500 /problems/internal-error`

## Alarms (0033)

### GET /api/alarms

Query params:
- status: active | acked | cleared (default active)
- severity: warn | alarm
- object_id: uuid
- from, to: RFC3339 (for status=cleared)
- limit (default 50, max 500)
- offset (default 0)

Response fields:
- items[]: tag_id, tag_name, device_id, device_name, object_id, object_name, state, value, quality, entered_at, last_seen_at, acked, ack
- total, offset, limit

### GET /api/alarms/{tagId}

Returns:
- current alarm state record
- last 50 events for tag

### POST /api/alarms/{tagId}/acknowledge

Body:
- note?: string|null

Errors:
- /errors/alarms/not-found (404)
- /errors/alarms/not-active (409)
- /errors/alarms/already-acked (409)
- /errors/alarms/invalid-request (400)

### WebSocket topic `alarms`

Subscribe:
- {"t":"subscribe","topics":["alarms"]}

Server pushes:
- alarms_snapshot: active unacked alarms snapshot
- alarm: raised/cleared/acked event

### Kafka topic

Producer publishes alarm events to `alarms.events` (message key = tag_id).

## Notifier (debug, 0035)

### GET /api/notifier/chats

Debug endpoint for checking Telegram chat bindings used by notifier routing.

Query params:
- enabled: bool
- role: string
- object_id: uuid

Response:
- items[]: id, chat_id, title, role, object_id, severity_min, enabled
- total

## Authorization (0036)

All `/api/*` endpoints require JWT and role checks.

- `401` => missing/invalid/expired token
- `403` => token is valid, but role is not allowed for endpoint

Role mapping:
- `GET /api/trends*`, `GET /api/tags`, `GET /api/objects*`, `GET /api/diagrams*`, `GET /api/alarms*` => `operator|engineer|admin`
- `POST /api/alarms/{tagId}/acknowledge` => `operator|admin`
- `GET /api/notifier/chats` => `engineer|admin`
- `GET /api/ws` (upgrade) => `operator|engineer|admin`

## Bulk acknowledge

`POST /api/alarms/acknowledge`

Request:

```json
{
  "items": [
    { "tag_id": "<uuid>" }
  ],
  "note": "РџСЂРёРЅСЏС‚Рѕ РІ СЂР°Р±РѕС‚Сѓ"
}
```

- `items`: РѕС‚ 1 РґРѕ 200 СѓРЅРёРєР°Р»СЊРЅС‹С… `tag_id` (РґСѓР±Р»РёРєР°С‚С‹ РІ С‚РµР»Рµ С‚РёС…Рѕ РґРµРґСѓРїР»РёС†РёСЂСѓСЋС‚СЃСЏ).
- `note`: РѕРїС†РёРѕРЅР°Р»СЊРЅРѕ, РґРѕ 500 СЃРёРјРІРѕР»РѕРІ.

Response body:

```json
{
  "items": [
    { "tag_id": "<uuid>", "status": "acked", "state": "hi" },
    { "tag_id": "<uuid>", "status": "not_active" }
  ],
  "acked_at": "2026-05-01T10:01:00.123Z",
  "actor_id": "user-1",
  "success_n": 1,
  "failed_n": 1
}
```

HTTP status:

- `200 OK`: РІСЃРµ СЌР»РµРјРµРЅС‚С‹ acked.
- `207 Multi-Status`: СЃРјРµС€Р°РЅРЅС‹Р№ СЂРµР·СѓР»СЊС‚Р°С‚.
- `409 Conflict`: РЅРё РѕРґРёРЅ СЌР»РµРјРµРЅС‚ РЅРµ acked.
- `422 Unprocessable Entity`: `/errors/alarms/bulk-size`, `/errors/alarms/note-too-long`.

WS:

- РќР° СѓСЃРїРµС€РЅС‹Рµ СЌР»РµРјРµРЅС‚С‹ СЃРµСЂРІРµСЂ РѕС‚РїСЂР°РІР»СЏРµС‚ РѕРґРёРЅ frame `{"t":"alarms_batch","events":[...]}` РІ topic `alarms`.
