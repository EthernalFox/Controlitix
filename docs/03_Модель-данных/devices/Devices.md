# devices.devices

Назначение: устройства (PLC, RTU, IP-устройства), источники данных. Устройство может существовать независимо от объекта мониторинга (пул устройств) и назначаться/отвязываться по необходимости.

Колонки
- id: uuid, PK
- object_id: uuid, **NULL**, FK -> `public.objects(id)` — nullable, устройство может быть не привязано к объекту
- type_id: int, NOT NULL, FK -> `devices.device_type(id)`
- name: text, NOT NULL
- description: text, NULL
- deleted_at: timestamptz, NULL — soft delete
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- N:1 с `public.objects` (по `devices.object_id`) — опциональная связь.
- 1:1 с `devices.devices_params` (по `device_id`).
- 1:N с `tags.tags` (по `tags.device_id`).

Индексы/ограничения
- PK(id)
- IDX(object_id)
- IDX(type_id)
- UNIQUE(object_id, name) WHERE deleted_at IS NULL — имя уникально внутри объекта
