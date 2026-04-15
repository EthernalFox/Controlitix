# tags.data_types

Назначение: справочник типов данных тегов.

Колонки
- id: int, PK
- type: text, NOT NULL, UNIQUE (например: bool, int16, uint16, float32, string)

Индексы/ограничения
- PK(id)
- UNIQUE(type)

