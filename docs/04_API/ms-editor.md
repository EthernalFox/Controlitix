# ms-editor API

Назначение: конфигурация объектов, устройств, тегов, мнемосхем, фигур и публикация изменений.

Точки API (набросок)

Объекты
- `POST /objects` - создать объект мониторинга
- `GET /objects/{id}` - получить объект
- `PATCH /objects/{id}` - обновить объект
- `DELETE /objects/{id}` - удалить объект
- `GET /objects/{id}/devices` - список устройств объекта

Устройства
- `POST /objects/{id}/devices` - создать устройство объекта
- `GET /devices/{id}` - получить устройство
- `PATCH /devices/{id}` - обновить устройство
- `DELETE /devices/{id}` - удалить устройство
- `POST /devices/{id}/config` - задать конфигурацию устройства

Теги
- `POST /devices/{id}/tags` - создать тег устройства
- `GET /devices/{id}/tags` - список тегов устройства
- `GET /tags/{id}` - получить тег
- `PATCH /tags/{id}` - обновить тег
- `DELETE /tags/{id}` - удалить тег
- `GET /device-types` - список типов устройств
- `GET /data-types` - список типов данных
- `GET /units` - список единиц измерения

Мнемосхемы (diagrams)
- `POST /objects/{id}/diagrams` - создать мнемосхему объекта
- `GET /diagrams/{id}` - получить мнемосхему
- `PATCH /diagrams/{id}` - обновить мнемосхему
- `DELETE /diagrams/{id}` - удалить мнемосхему
- `POST /diagrams/{id}/publish` - публикация текущего состояния

Фигуры
- `POST /diagrams/{id}/figures` - добавить фигуры
- `PATCH /figures/{id}` - обновить фигуру/параметры
- `DELETE /figures/{id}` - удалить фигуру

DTO (MVP, набросок)

Объект мониторинга
- Запрос: `name` (обязательно), `description` (опционально).
- Ответ: `id`, `name`, `description`, `created_at`, `updated_at`.

Устройство
- Запрос: `object_id` берётся из пути `/objects/{id}/devices`; поля `name`, `description`, `type`, `settings` (JSONB).
- Ответ: `id`, `object_id`, `type`, `name`, `description`, `settings`, `created_at`, `updated_at`.

Тег
- Запрос: `device_id` берётся из пути `/devices/{id}/tags`; поля `name`, `description`, `address`, `data_type_id`, `unit_id`, `register_type`, `scaling`, `setpoints`.
- Ответ: `id`, `device_id`, `name`, `description`, `params`, `created_at`, `updated_at`.

Мнемосхема
- Запрос: `name` (опционально), `description` (опционально). `object_id` берется из пути `/objects/{id}/diagrams`.
- Ответ: `id`, `object_id`, `name`, `description`, `published_at`, `created_at`, `updated_at`.

Фигура
- Запрос: `type` (обязательно), `params` (JSONB), `tag_id` (опционально).
- Ответ: `id`, `scheme_id`, `tag_id`, `params`, `created_at`, `updated_at`.

Связанные таблицы и маппинг
- Объекты: `public.objects` -> `../../03_Модель-данных/public/Objects.md`
- Устройства: `devices.devices` -> `../../03_Модель-данных/devices/Devices.md`
- Параметры устройств: `devices.devices_params` -> `../../03_Модель-данных/devices/Devices_params.md`
- Типы устройств: `devices.device_type` -> `../../03_Модель-данных/devices/Device_type.md`
- Теги: `tags.tags` -> `../../03_Модель-данных/tags/Tags.md`
- Параметры тегов: `tags.tag_params` -> `../../03_Модель-данных/tags/Tag_params.md`
- Справочники: `tags.data_types` -> `../../03_Модель-данных/tags/Data_types.md`, `tags.units` -> `../../03_Модель-данных/tags/Units.md`
- Расширения тегов: `tags.tag_setpoints` -> `../../03_Модель-данных/tags/Tag_setpoints.md`, `tags.tag_scaling` -> `../../03_Модель-данных/tags/Tag_Scaling.md`
- Мнемосхемы: `public.mimic` -> `../../03_Модель-данных/public/Mimic.md`
- Фигуры: `public.figures` -> `../../03_Модель-данных/public/Figures.md`
- Параметры фигур: `public.figure_params` -> `../../03_Модель-данных/public/Figure_params.md`

Расширить (planned): версии, черновики, права доступа, схемы DTO и детальные коды ошибок.
