# public.figure_params

Назначение: параметры/настройки отображения фигуры (атрибуты, логика, выражения), хранение в JSONB.

Колонки
- figure_id: uuid, PK, FK → `public.figures(id)`
- params: JSONB, NOT NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:1 с `public.figures` (PK=FK).

Индексы/ограничения
- PK(figure_id)
- GIN(params) — рекомендовано для поиска по JSONB.

