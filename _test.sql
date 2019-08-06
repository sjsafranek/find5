

SELECT json_agg(c)
FROM (
    SELECT
        (EXTRACT(epoch FROM created_at) / EXTRACT(epoch FROM INTERVAL '5 sec'))::INTEGER AS bucket,
        json_agg(
            json_build_object(
                'location_id',
                measurements.location_id,
                'key',
                measurements.key,
                'value',
                measurements.value
            )
        ) AS measurements
    FROM measurements
    WHERE
        location_id IS NOT NULL
    GROUP BY bucket
) c;



SELECT json_agg(c)
FROM (
    SELECT
        (EXTRACT(epoch FROM created_at) / EXTRACT(epoch FROM INTERVAL '5 sec'))::INTEGER AS bucket,
        location_id,
        json_agg(
            json_build_object(
                'key',
                measurements.key,
                'value',
                measurements.value
            )
        ) AS measurements
    FROM measurements
    WHERE
        location_id IS NOT NULL
    GROUP BY bucket, location_id
) c;




SELECT json_agg(l) FROM (
    SELECT
        location_buckets.location_id,
        json_agg(location_buckets.bucket) AS buckets
    FROM (
        SELECT
            buckets.location_id,
            json_build_object(
                'bucket',
                buckets.bucket,
                'measurements',
                json_agg(buckets.measurements)->0
            ) AS bucket
        FROM (
            SELECT
                (EXTRACT(epoch FROM measurements.created_at) / EXTRACT(epoch FROM INTERVAL '5 sec'))::INTEGER AS bucket,
                measurements.location_id,
                json_agg(
                    json_build_object(
                        'sensor_id',
                        sensors.id,
                        'key',
                        measurements.key,
                        'value',
                        measurements.value
                    )
                ) AS measurements
            FROM measurements
            INNER JOIN sensors
                ON sensors.id = measurements.sensor_id
                AND sensors.is_deleted = false
            INNER JOIN devices
                ON devices.id = sensors.device_id
                AND devices.username = 'admin'
            WHERE
                measurements.location_id IS NOT NULL
            GROUP BY bucket, location_id
        ) AS buckets
        GROUP BY buckets.bucket, buckets.location_id
    ) location_buckets
    GROUP BY location_buckets.location_id
) AS l;









WITH measurements AS (
        SELECT
            (EXTRACT(epoch FROM measurements.created_at) / EXTRACT(epoch FROM INTERVAL '5 sec'))::INTEGER AS bucket,
            measurements.location_id,
            sensors.id AS sensor_id,
            json_agg(
                json_build_object(
                    'key',
                    measurements.key,
                    'value',
                    measurements.value
                )
            ) AS measurements
        FROM measurements
        INNER JOIN sensors
            ON sensors.id = measurements.sensor_id
            AND sensors.is_deleted = false
        INNER JOIN devices
            ON devices.id = sensors.device_id
            AND devices.username = 'admin'
        WHERE
            measurements.location_id IS NOT NULL
        GROUP BY bucket, measurements.location_id, sensors.id
    ),
    buckets AS (
        SELECT
            measurements.location_id,
            measurements.sensor_id,
            json_build_object(
                'bucket_id',
                measurements.bucket,
                'measurements',
                json_agg(measurements.measurements)->0
            ) AS bucket
        FROM measurements
        GROUP BY measurements.bucket, measurements.location_id, measurements.sensor_id
    ),
    location_buckets AS (
        SELECT
            buckets.sensor_id,
            buckets.location_id,
            json_agg(buckets.bucket) AS buckets
        FROM buckets
        GROUP BY buckets.location_id, buckets.sensor_id
    ),
    sensor_locations AS (
        SELECT
            location_buckets.location_id,
            json_agg(
                json_build_object(
                'sensor_id',
                location_buckets.sensor_id,
                'buckets',
                location_buckets.buckets
                )
            ) AS sensors
        FROM location_buckets
        GROUP BY location_buckets.location_id
    )

SELECT json_agg(c) FROM (
    SELECT
        location_id,
        sensors
    FROM sensor_locations
) c;
