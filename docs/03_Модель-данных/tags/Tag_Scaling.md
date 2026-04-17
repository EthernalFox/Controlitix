# tags.tag_scaling

Назначение: параметры масштабирования для приведения «сырых» значений к инженерным. Поддерживает два метода, которые могут применяться совместно: линейная интерполяция и множитель+смещение.

Порядок обработки: сначала интерполяция (если задана), затем factor + offset.

Колонки
- param_id: uuid, PK, FK → `tags.tag_params(id)` ON DELETE CASCADE
- raw_min: double precision, NULL — нижняя граница сырого диапазона (для интерполяции)
- raw_max: double precision, NULL — верхняя граница сырого диапазона
- eng_min: double precision, NULL — нижняя граница инженерного диапазона
- eng_max: double precision, NULL — верхняя граница инженерного диапазона
- factor: double precision, NULL — множитель (по умолчанию 1.0)
- "offset": double precision, NULL — смещение (по умолчанию 0.0)
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Валидация:
- Если используется интерполяция, все четыре поля (raw_min, raw_max, eng_min, eng_max) должны быть заданы, raw_min ≠ raw_max.
- factor не должен быть равен 0 (если задан).

Связи
- 1:1 с `tags.tag_params` (PK=FK).

Индексы/ограничения
- PK(param_id)

