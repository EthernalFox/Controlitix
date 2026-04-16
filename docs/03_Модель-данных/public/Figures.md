# public.figures

Назначение: графические элементы мнемосхем (статические/динамические).

Колонки
- id: uuid, PK
- scheme_id: uuid, NOT NULL, FK → `public.mimic(id)`
- tag_id: uuid, NULL, FK → `tags.tags(id)` — если фигура привязана к тегу
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL
- type: хранится логически в связке с `figure_params.params.type` и на уровне API, фактический whitelist типов описан в `../../06_Дизайн/Figure-Types.md`

Связи
- N:1 с `public.mimic`.
- 1:1 с `public.figure_params` (через `figure_params.figure_id`).

Индексы/ограничения
- PK(id)
- IDX(scheme_id), IDX(tag_id)

Примечание
- В исходной диаграмме `scheme_id` отмечен как int и `tagID` как UUID. Исправлено на `scheme_id uuid`, `tag_id uuid` (snake_case).
- На уровне БД используется историческое имя `scheme_id`, но в публичном HTTP API рекомендуется использовать `diagram_id`, чтобы совпадать с остальной терминологией сервиса.
