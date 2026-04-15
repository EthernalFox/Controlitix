# 3. Модель данных

Источник: `ControlitixDB.drawio` (архивная версия). Разделение на схемы: `public`, `devices`, `tags`, `history`, `alarms`, `auth`.

Разделение на схемы внутри одной БД — сознательный выбор, заложенный под будущий переход в SaaS. Подробнее — в [ADR-0003](../08_ADR/ADR-0003-postgres-schemas.md). **Кросс-схемные FK разрешены** (PostgreSQL поддерживает их штатно).

## Схемы и таблицы

- 3.1 `public` — объекты мониторинга, мнемосхемы и HMI
  - 3.1.1 [objects](public/Objects.md)
  - 3.1.2 [mimic](public/Mimic.md)
  - 3.1.3 [figures](public/Figures.md)
  - 3.1.4 [figure_params](public/Figure_params.md)
- 3.2 `devices` — устройства и их параметры
  - 3.2.1 [devices](devices/Devices.md)
  - 3.2.2 [device_type](devices/Device_type.md)
  - 3.2.3 [devices_params](devices/Devices_params.md)
- 3.3 `tags` — теги и атрибуты
  - 3.3.1 [tags](tags/Tags.md)
  - 3.3.2 [tag_params](tags/Tag_params.md)
  - 3.3.3 [data_types](tags/Data_types.md)
  - 3.3.4 [units](tags/Units.md)
  - 3.3.5 [tag_setpoints](tags/Tag_setpoints.md)
  - 3.3.6 [tag_scaling](tags/Tag_Scaling.md)
- 3.4 `history` — архив значений тегов (TimescaleDB hypertables и continuous aggregates) — TBD
- 3.5 `alarms` — тревоги, события тревог, квитирования — TBD
- 3.6 `auth` — пользователи, роли, токены, telegram_chats, audit_log — TBD

## Принятые соглашения

- **Типы ключей:** сущностные PK/FK — `uuid`; справочники (`device_type`, `data_types`, `units`) — `int`.
- **Аудит полей:** `created_at`, `updated_at` (snake_case) во всех сущностных таблицах.
- **1:1 таблицы** оформлены как `PK = FK`: `devices_params(device_id)`, `figure_params(figure_id)`, `tag_setpoints(param_id)`, `tag_scaling(param_id)`.
- **Кросс-схемные FK:** разрешены. Конкретные связи:
  - `devices.devices.object_id → public.objects.id` — FK.
  - `tags.tags.device_id → devices.devices.id` — FK.
  - `history.*` — **без FK** на `tags.tags` (высокая нагрузка, hypertable, целостность поддерживается приложением).
  - `alarms.*.tag_id` — **без FK** (логическая связь).
  - `public.figures.tag_id` — **без FK** (логическая связь).
- **Мягкое удаление:** `deleted_at timestamptz null` применяется в `public.objects`, `public.mimic`, `public.figures`. Уникальные ограничения оформляются частичными индексами `where deleted_at is null`.
- **Физическое удаление:** в MVP физически удаляются записи `devices.*` и `tags.*` (каскад через FK).

