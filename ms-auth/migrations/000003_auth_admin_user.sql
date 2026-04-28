-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    v_username text := current_setting('ms_auth.initial_admin_username', true);
    v_hash text := current_setting('ms_auth.initial_admin_password_hash', true);
    v_email text := current_setting('ms_auth.initial_admin_email', true);
    v_user_id uuid;
BEGIN
    IF v_username IS NULL OR v_username = '' OR v_hash IS NULL OR v_hash = '' THEN
        RAISE NOTICE 'ms_auth: initial admin vars not set, skipping';
        RETURN;
    END IF;

    INSERT INTO auth.users (subject, source_id, username, display_name, email, password_hash, is_active)
    VALUES (
        'local:' || gen_random_uuid()::text,
        1,
        v_username,
        'Administrator',
        NULLIF(v_email, ''),
        v_hash,
        true
    )
    ON CONFLICT (source_id, lower(username)) DO NOTHING
    RETURNING id INTO v_user_id;

    IF v_user_id IS NULL THEN
        SELECT id INTO v_user_id
        FROM auth.users
        WHERE source_id = 1 AND lower(username) = lower(v_username);
    END IF;

    INSERT INTO auth.user_roles (user_id, role_code)
    VALUES (v_user_id, 'admin')
    ON CONFLICT DO NOTHING;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    v_username text := current_setting('ms_auth.initial_admin_username', true);
BEGIN
    IF v_username IS NULL OR v_username = '' THEN
        RETURN;
    END IF;

    DELETE FROM auth.users
    WHERE source_id = 1 AND lower(username) = lower(v_username);
END $$;
-- +goose StatementEnd
