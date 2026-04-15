# tags.tag_params

Назначение: техническая спецификация чтения/интерпретации значения тега.

Колонки
- id: uuid, PK
- tag_id: uuid, NOT NULL, FK → `tags.tags(id)`
- data_type_id: int, NOT NULL, FK → `tags.data_types(id)`
- unit_id: int, NULL, FK → `tags.units(id)`
- address: text, NULL — адрес в устройстве (например, Modbus регистр, OID для SNMP)
- scale: numeric, NULL — масштабирование простым коэффициентом (если не используется `tag_scaling`)
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- N:1 с `tags.tags`.
- 1:1 расширения: `tags.tag_setpoints(param_id)` и `tags.tag_scaling(param_id)`.

Индексы/ограничения
- PK(id)
- IDX(tag_id)

