# tags.tag_scaling

Назначение: параметры масштабирования для приведения «сырых» значений к инженерным.

Колонки
- param_id: uuid, PK, FK → `tags.tag_params(id)`
- min_value: double precision, NULL
- max_value: double precision, NULL
- factor: double precision, NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:1 с `tags.tag_params` (PK=FK).

Индексы/ограничения
- PK(param_id)

