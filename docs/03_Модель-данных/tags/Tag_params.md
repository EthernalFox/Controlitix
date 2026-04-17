# tags.tag_params

Назначение: техническая спецификация чтения/интерпретации значения тега.

Колонки
- id: uuid, PK
- tag_id: uuid, NOT NULL, FK → `tags.tags(id)` ON DELETE CASCADE
- data_type_id: int, NOT NULL, FK → `tags.data_types(id)`
- unit_id: int, NULL, FK → `tags.units(id)`
- address: jsonb, NULL — адрес в устройстве, структура зависит от типа устройства (см. ниже)
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Формат address по типу устройства:

Modbus (modbus_rtu, modbus_tcp):
```json
{"register_type": "holding", "address": 40001}
```
`register_type`: `"holding"` | `"input"` | `"coil"` | `"discrete"`

SNMP (snmp_v1, snmp_v2c, snmp_v3):
```json
{"oid": "1.3.6.1.2.1.1.1.0"}
```

Связи
- N:1 с `tags.tags`.
- 1:1 расширения: `tags.tag_setpoints(param_id)` и `tags.tag_scaling(param_id)`.

Индексы/ограничения
- PK(id)
- IDX(tag_id)

