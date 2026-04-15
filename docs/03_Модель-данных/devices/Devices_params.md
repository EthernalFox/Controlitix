# devices.devices_params

Назначение: параметры подключения/опроса устройства (Modbus TCP/RTU, SNMP и др.).

Колонки
- device_id: uuid, PK, FK → `devices.devices(id)`
- settings: JSONB, NOT NULL
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Связи
- 1:1 с `devices.devices` (PK=FK).

Индексы/ограничения
- PK(device_id)
- GIN(settings) — рекомендовано для поиска по JSONB.

Примечание
- В исходной диаграмме встречался `id (int)`; нормализовано до PK=FK по `device_id (uuid)`.

