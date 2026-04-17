# tags.units

Назначение: справочник единиц измерения с категоризацией для удобного поиска и фильтрации.

Колонки
- id: serial, PK
- name: text, NOT NULL — полное название (например: Kilowatt, Celsius, Pascal)
- symbol: text, NOT NULL, UNIQUE — краткое обозначение (например: kW, °C, Pa)
- category: text, NOT NULL — категория (electrical, temperature, pressure, humidity, percentage, frequency, time)

Категории MVP:
- `electrical` — V, mV, kV, A, mA, W, kW, MW, kWh, MWh, VA, kVA, Hz, Ω, kΩ
- `temperature` — °C, °F, K
- `pressure` — Pa, kPa, MPa, bar, mbar
- `humidity` — %RH
- `percentage` — %
- `frequency` — RPM
- `time` — s, ms, min, h

Индексы/ограничения
- PK(id)
- UNIQUE(symbol)
- IDX(category) — для фильтрации `GET /units?category=electrical`

