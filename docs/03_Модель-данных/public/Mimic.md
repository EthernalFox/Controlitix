# public.mimic

Назначение: мнемосхемы объектов (HMI-схемы), с версиями и публикацией (в модели хранения - одна текущая версия).

Колонки
- id: uuid, PK
- object_id: uuid, NOT NULL, FK -> `public.objects(id)`
- name: text, NULL (если требуется именование схемы)
- description: text, NULL
- published_at: timestamptz, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- N:1 с `public.objects`.
- 1:N с `public.figures` (`figures.scheme_id` -> `mimic.id`).

Индексы/ограничения
- PK(id)
- IDX(object_id)

Примечание
- В исходной диаграмме `object_id` указан как int. Исправлено на uuid для согласованности.
