
DROP VIEW IF EXISTS users_view CASCADE;

CREATE OR REPLACE VIEW users_view AS (
    SELECT
        *,
        json_build_object(
            'email', email,
            'username', username,
            'apikey', apikey,
            'secret_token', secret_token,
            'is_active', is_active,
            'is_deleted', is_deleted,
            'created_at', to_char(created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
            'updated_at', to_char(updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"')
        ) AS user_json
    FROM users
);
