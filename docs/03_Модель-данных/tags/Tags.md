# tags.tags

Назначение: теги — измерения/состояния, получаемые с устройств.

Колонки
- id: uuid, PK
- device_id: uuid, NOT NULL, FK → `devices.devices(id)`
- name: text, NOT NULL
- description: text, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:N с `tags.tag_params` (по `tag_params.tag_id`). В модели хранения допускается 1:1 (одна активная запись), но может быть расширено до версионирования.
- N:1 с `devices.devices`.

Индексы/ограничения
- PK(id)
- IDX(device_id), UNIQUE(device_id, name) — рекомендуется для уникальности имён тегов в рамках устройства.

