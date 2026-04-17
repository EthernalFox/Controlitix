# tags.tags

Назначение: теги — измерения/состояния, получаемые с устройств.

Колонки
- id: uuid, PK
- device_id: uuid, NOT NULL, FK → `devices.devices(id)`
- name: text, NOT NULL
- description: text, NULL
- deleted_at: timestamptz, NULL — soft delete
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:1 с `tags.tag_params` (по `tag_params.tag_id`). ON DELETE CASCADE.
- N:1 с `devices.devices`.

Индексы/ограничения
- PK(id)
- IDX(device_id)
- UNIQUE(device_id, name) WHERE deleted_at IS NULL — имя тега уникально в рамках устройства

