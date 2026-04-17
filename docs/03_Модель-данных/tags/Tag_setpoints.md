# tags.tag_setpoints

Назначение: уставки/пороги для тега. Стандартные SCADA-уровни: LoLo < Lo < Hi < HiHi.

Колонки
- param_id: uuid, PK, FK → `tags.tag_params(id)` ON DELETE CASCADE
- lolo: double precision, NULL — аварийный нижний порог
- lo: double precision, NULL — предупредительный нижний порог
- hi: double precision, NULL — предупредительный верхний порог
- hihi: double precision, NULL — аварийный верхний порог
- created_at: timestamptz, NOT NULL
- updated_at: timestamptz, NOT NULL

Валидация: если несколько уставок заданы, порядок строгий: lolo < lo < hi < hihi.

Связи
- 1:1 с `tags.tag_params` (PK=FK).

Индексы/ограничения
- PK(param_id)

