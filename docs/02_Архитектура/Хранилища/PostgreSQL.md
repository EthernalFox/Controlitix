# PostgreSQL

Назначение
- Хранение конфигураций: схемы `public`, `devices`, `tags`.
- Историзация значений (по agreed варианту): отдельные таблицы/расширение (например, TimescaleDB) — TBA.

Схемы и таблицы
- Обзор: см. `../../03_Модель-данных/README.md`.

Краткое саммари схем
- public — объекты и HMI:
  - `objects` → `../../03_Модель-данных/public/Objects.md`
  - `mimic` → `../../03_Модель-данных/public/Mimic.md`
  - `figures` → `../../03_Модель-данных/public/Figures.md`
  - `figure_params` → `../../03_Модель-данных/public/Figure_params.md`
- devices — устройства и их параметры:
  - `devices` → `../../03_Модель-данных/devices/Devices.md`
  - `device_type` → `../../03_Модель-данных/devices/Device_type.md`
  - `devices_params` → `../../03_Модель-данных/devices/Devices_params.md`
- tags — теги и атрибуты:
  - `tags` → `../../03_Модель-данных/tags/Tags.md`
  - `tag_params` → `../../03_Модель-данных/tags/Tag_params.md`
  - `data_types` → `../../03_Модель-данных/tags/Data_types.md`
  - `units` → `../../03_Модель-данных/tags/Units.md`
  - `tag_setpoints` → `../../03_Модель-данных/tags/Tag_setpoints.md`
  - `tag_scaling` → `../../03_Модель-данных/tags/Tag_Scaling.md`

Требования и политики
- Транзакционная целостность конфигураций.
- Индексы по FK и по полям аудита `created_at`, `updated_at` (по необходимости).
- Резервное копирование и политика retention (описать в «Порядок внедрения»).
