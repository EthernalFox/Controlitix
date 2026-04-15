# public.figures

Назначение: графические элементы мнемосхем (статические/динамические).

Колонки
- id: uuid, PK
- scheme_id: uuid, NOT NULL, FK → `public.mimic(id)`
- tag_id: uuid, NULL, FK → `tags.tags(id)` — если фигура привязана к тегу
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- N:1 с `public.mimic`.
- 1:1 с `public.figure_params` (через `figure_params.figure_id`).

Индексы/ограничения
- PK(id)
- IDX(scheme_id), IDX(tag_id)

Примечание
- В исходной диаграмме `scheme_id` отмечен как int и `tagID` как UUID. Исправлено на `scheme_id uuid`, `tag_id uuid` (snake_case).

