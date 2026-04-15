# ms-poll — опрос устройств

Назначение
- Читает конфигурацию из PostgreSQL (`devices.*`, `tags.*`) и опрашивает устройства (Modbus TCP/RTU, SNMP).
- Публикует текущие значения в Kafka (топик `ms-poll`) и/или Redis pub/sub.

Границы ответственности
- Не изменяет конфигурацию; источник правды — `ms-editor`.
- Планировщик опроса (scan rate), backoff/retry, нормализация (тип/скейлинг/quality).

Интерфейсы
- Вход: Kafka `config.changed`, PostgreSQL конфиг.
- Выход: Kafka `ms-poll` (payload: `tag_id`, `value`, `ts`, `quality`, `meta`), `raw.readings` (опц.).

Зависимости
- PostgreSQL, Kafka, Redis (опц.).

НФТ
- Метрики по протоколам (RPS/ошибки/latency), трейсинг, логи.
