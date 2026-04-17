# devices.device_type

Назначение: справочник типов устройств (протоколов подключения).

Колонки
- id: serial, PK
- name: text, NOT NULL, UNIQUE

Seed-данные MVP:
- `modbus_rtu`
- `modbus_tcp`
- `snmp_v1`
- `snmp_v2c`
- `snmp_v3`

Индексы/ограничения
- PK(id)
- UNIQUE(name)

