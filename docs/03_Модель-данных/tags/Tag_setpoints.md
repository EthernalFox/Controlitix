# tags.tag_setpoints

Назначение: уставки/пороги для тега.

Колонки
- param_id: uuid, PK, FK → `tags.tag_params(id)`
- lolo: double precision, NULL
- lohi: double precision, NULL
- hilo: double precision, NULL
- hihi: double precision, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:1 с `tags.tag_params` (PK=FK).

Индексы/ограничения
- PK(param_id)

