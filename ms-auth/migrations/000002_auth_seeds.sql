-- +goose Up
INSERT INTO auth.identity_sources (id, type, name, config, is_enabled)
VALUES (1, 'local', 'Local database', '{}'::jsonb, true)
ON CONFLICT (id) DO NOTHING;

SELECT setval(
    'auth.identity_sources_id_seq',
    GREATEST((SELECT MAX(id) FROM auth.identity_sources), 1)
);

INSERT INTO auth.roles (code, name, description)
VALUES
    ('admin', 'Администратор', 'Полный доступ, управление пользователями и ТУЗ'),
    ('engineer', 'Инженер', 'CRUD инженерной конфигурации (objects, devices, tags, diagrams)'),
    ('operator', 'Оператор', 'Просмотр мнемосхем, трендов, квитирование тревог')
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DELETE FROM auth.roles WHERE code IN ('admin', 'engineer', 'operator');
DELETE FROM auth.identity_sources WHERE id = 1;
