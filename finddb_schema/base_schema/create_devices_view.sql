
DROP VIEW IF EXISTS devices_view CASCADE;

CREATE OR REPLACE VIEW devices_view AS (
    SELECT
        devices.*,
        json_build_object(
            'id', devices.id,
            'name', devices.name,
            'type', devices.type,
            'username', devices.username,
            'is_active', devices.is_active,
            'is_deleted', devices.is_deleted,
            'created_at', to_char(devices.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
            'updated_at', to_char(devices.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
            'sensors',
            (
                SELECT json_agg(s) FROM (
                    SELECT
                        sensors.id,
                        sensors.name,
                        sensors.type,
                        sensors.device_id,
                        sensors.is_active,
                        sensors.is_deleted,
                        to_char(sensors.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as created_at,
                        to_char(sensors.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as updated_at
                    FROM sensors AS sensors
                    WHERE sensors.device_id=devices.id
                    AND sensors.is_deleted = false
                ) s
            )
        ) AS devices_json
    FROM devices AS devices
);
