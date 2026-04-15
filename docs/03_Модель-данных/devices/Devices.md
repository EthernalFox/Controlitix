# devices.devices

Назначение: устройства (PLC, RTU, IP-устройства), источники данных.

Колонки
- id: uuid, PK
- object_id: uuid, NOT NULL, FK -> `public.objects(id)`
- type: int, NOT NULL, FK -> `devices.device_type(id)`
- name: text, NOT NULL
- description: text, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- N:1 с `public.objects` (по `devices.object_id`).
- 1:1 с `devices.devices_params` (по `device_id`).
- 1:N с `tags.tags` (по `tags.device_id`).

Индексы/ограничения
- PK(id)
- IDX(object_id)
- IDX(type), UNIQUE(name) - опционально, если требуется уникальность имён.
