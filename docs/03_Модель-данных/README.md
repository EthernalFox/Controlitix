# 3. Модель данных

Источник: `ControlitixDB.drawio`. Схемы: `public`, `devices`, `tags`.

- 3.1 public
  - 3.1.1 [objects](public/Objects.md)
  - 3.1.2 [mimic](public/Mimic.md)
  - 3.1.3 [figures](public/Figures.md)
  - 3.1.4 [figure_params](public/Figure_params.md)
- 3.2 devices
  - 3.2.1 [devices](devices/Devices.md)
  - 3.2.2 [device_type](devices/Device_type.md)
  - 3.2.3 [devices_params](devices/Devices_params.md)
- 3.3 tags
  - 3.3.1 [tags](tags/Tags.md)
  - 3.3.2 [tag_params](tags/Tag_params.md)
  - 3.3.3 [data_types](tags/Data_types.md)
  - 3.3.4 [units](tags/Units.md)
  - 3.3.5 [tag_setpoints](tags/Tag_setpoints.md)
  - 3.3.6 [tag_scaling](tags/Tag_Scaling.md)

Принятые исправления:
- Унификация типов: сущностные PK/FK — `uuid`; справочники (`device_type`, `data_types`, `units`) — `int`.
- Единый стиль аудита: `created_at`, `updated_at` (snake_case) во всех сущностных таблицах.
- `tag_params.unit_id` — `int` (FK → `units.id`).
- 1:1 таблицы оформлены как PK=FK: `devices_params(device_id)`, `figure_params(figure_id)`, `tag_setpoints(param_id)`, `tag_scaling(param_id)`.

