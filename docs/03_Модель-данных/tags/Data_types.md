# tags.data_types

Назначение: справочник типов данных тегов.

Колонки
- id: serial, PK
- name: text, NOT NULL, UNIQUE

Seed-данные MVP:
- `bool`, `int8`, `uint8`, `int16`, `uint16`, `int32`, `uint32`, `float32`, `float64`, `string`

Индексы/ограничения
- PK(id)
- UNIQUE(name)

