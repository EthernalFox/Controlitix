# public.objects

Назначение: контейнеры мест мониторинга (ЦОД, узел и т. п.).

Колонки
- id: uuid, PK
- name: text, NOT NULL
- description: text, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:N с `public.mimic` (`mimic.object_id` -> `objects.id`).
- 1:N с `devices.devices` (`devices.object_id` -> `objects.id`).

Индексы/ограничения
- PK(id)
- Рекомендуется индекс по `created_at`.
