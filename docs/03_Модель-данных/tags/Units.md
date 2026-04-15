# tags.units

Назначение: справочник единиц измерения.

Колонки
- id: int, PK
- unit: text, NOT NULL, UNIQUE (например: °C, Pa, %, A)

Индексы/ограничения
- PK(id)
- UNIQUE(unit)

