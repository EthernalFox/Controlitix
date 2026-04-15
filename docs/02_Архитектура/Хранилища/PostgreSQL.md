# PostgreSQL + TimescaleDB

## Назначение

Единая база данных Controlitix. Используется **PostgreSQL с extension TimescaleDB** — отдельных инстансов не поднимаем. TimescaleDB отвечает за архив значений тегов (hypertables, continuous aggregates, retention policies). Решение зафиксировано в [ADR-0004](../../08_ADR/ADR-0004-timescaledb-history.md).

## Схемы

Шесть логических схем внутри одной БД (см. [ADR-0003](../../08_ADR/ADR-0003-postgres-schemas.md)):

| Схема | Назначение | Ответственный сервис |
|---|---|---|
| `public` | Объекты мониторинга, мнемосхемы, фигуры | ms-editor |
| `devices` | Устройства и параметры подключения | ms-editor |
| `tags` | Теги, параметры, справочники, уставки, скейлинг | ms-editor |
| `history` | TimescaleDB hypertables, continuous aggregates | ms-viewer (write), ms-viewer (read) |
| `alarms` | Тревоги, события, квитирования | ms-viewer |
| `auth` | Пользователи, роли, токены, telegram_chats, audit | ms-auth |

**Кросс-схемные FK разрешены** — см. [ADR-0003](../../08_ADR/ADR-0003-postgres-schemas.md).

## Краткое саммари схем

- **public** — объекты и HMI:
  - `objects` → [`../../03_Модель-данных/public/Objects.md`](../../03_Модель-данных/public/Objects.md)
  - `mimic` → [`../../03_Модель-данных/public/Mimic.md`](../../03_Модель-данных/public/Mimic.md)
  - `figures` → [`../../03_Модель-данных/public/Figures.md`](../../03_Модель-данных/public/Figures.md)
  - `figure_params` → [`../../03_Модель-данных/public/Figure_params.md`](../../03_Модель-данных/public/Figure_params.md)
- **devices** — устройства и их параметры:
  - `devices` → [`../../03_Модель-данных/devices/Devices.md`](../../03_Модель-данных/devices/Devices.md)
  - `device_type` → [`../../03_Модель-данных/devices/Device_type.md`](../../03_Модель-данных/devices/Device_type.md)
  - `devices_params` → [`../../03_Модель-данных/devices/Devices_params.md`](../../03_Модель-данных/devices/Devices_params.md)
- **tags** — теги и атрибуты:
  - `tags` → [`../../03_Модель-данных/tags/Tags.md`](../../03_Модель-данных/tags/Tags.md)
  - `tag_params` → [`../../03_Модель-данных/tags/Tag_params.md`](../../03_Модель-данных/tags/Tag_params.md)
  - `data_types` → [`../../03_Модель-данных/tags/Data_types.md`](../../03_Модель-данных/tags/Data_types.md)
  - `units` → [`../../03_Модель-данных/tags/Units.md`](../../03_Модель-данных/tags/Units.md)
  - `tag_setpoints` → [`../../03_Модель-данных/tags/Tag_setpoints.md`](../../03_Модель-данных/tags/Tag_setpoints.md)
  - `tag_scaling` → [`../../03_Модель-данных/tags/Tag_Scaling.md`](../../03_Модель-данных/tags/Tag_Scaling.md)
- **history** — архив значений тегов (TimescaleDB) — детали в разделе 3.4 (TBD).
- **alarms** — тревоги и квитирования — детали в разделе 3.5 (TBD).
- **auth** — пользователи, токены, telegram_chats, audit — детали в разделе 3.6 (TBD).

## История значений (TimescaleDB)

Основные объекты схемы `history`:

- `history.tag_values_raw` — hypertable, `chunk_interval = 1 day`, partition key `ts`.
- `history.tag_values_1m` — continuous aggregate (1-минутные агрегаты: min/max/avg/count/last).
- **Retention policies:**
  - `tag_values_raw` — 30 дней.
  - `tag_values_1m` — 365 дней.
- **Compression policy:** chunks старше 7 дней сжимаются колоночным образом.

Маршрутизация запросов трендов в `ms-viewer` — окно ≤ 24 ч читается из raw, окно > 24 ч — из `tag_values_1m`.

## Требования и политики

- Транзакционная целостность конфигурационных данных (`ms-editor` открывает транзакции при сложных операциях).
- Индексы по FK; уникальные частичные индексы для таблиц с soft delete (`where deleted_at is null`).
- Миграции — Goose. Миграция TimescaleDB-hypertables и continuous aggregates прописывается явно.
- Резервное копирование:
  - Ежедневный `pg_dump` конфигурационных схем (`public`, `devices`, `tags`, `alarms`, `auth`) — быстро и безопасно.
  - Отдельная стратегия для `history` (hypertables): либо WAL-archive + базовый снапшот, либо `pg_dump` с учётом Timescale-оговорок. Детали — в [`../Развёртывание.md`](../Развёртывание.md) (TBD).
- Пользователи БД: отдельные роли на сервис (`ms_editor`, `ms_poll`, `ms_viewer`, `ms_auth`) с минимально необходимыми правами на «свои» схемы.

## Связанные документы

- [`../../03_Модель-данных/README.md`](../../03_Модель-данных/README.md)
- [`../../08_ADR/ADR-0003-postgres-schemas.md`](../../08_ADR/ADR-0003-postgres-schemas.md)
- [`../../08_ADR/ADR-0004-timescaledb-history.md`](../../08_ADR/ADR-0004-timescaledb-history.md)
