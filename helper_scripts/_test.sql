
WITH measurements_locations AS (
        SELECT
            measurements.key,
            COUNT(DISTINCT(measurements.sensor_id)) AS sensors,
            COUNT(measurements.*),
            MIN(measurements.value),
            MAX(measurements.value),
            STDDEV(measurements.value),
            AVG(measurements.value),
            MIN(measurements.created_at) AS first_timestamp,
            MAX(measurements.created_at) AS lastest_timestamp,
            locations.id AS location_id,
            locations.name AS location_name,
            ST_AsGeoJSON(locations.geom)::JSONB AS geometry
        FROM measurements
            INNER JOIN locations
                ON locations.id = measurements.location_id
        INNER JOIN sensors
            ON sensors.id = measurements.sensor_id
            AND sensors.is_deleted = false
        INNER JOIN devices
            ON devices.id = sensors.device_id
            ANd devices.username = 'admin'
            AND devices.is_deleted = false
        GROUP BY
            locations.id, measurements.key
    )

SELECT json_agg(c)
FROM (
    SELECT
        location_id,
        location_name,
        geometry,
        json_agg(
            json_build_object(
                'key',
                key,
                'sensors',
                sensors,
                'min',
                min,
                'max',
                max,
                'stddev',
                stddev,
                'mean',
                avg,
                'first_timestamp',
                first_timestamp,
                'lastest_timestamp',
                lastest_timestamp
            )
        ) AS scanners
    FROM
        measurements_locations
    GROUP BY
        location_id,
        location_name,
        geometry
) c;














-- GetByLocation

-- get count of mac addresses seen by device within time window


WITH device_locations AS (
    SELECT
        locations.id AS location_id,
        locations.name AS location_name,
        ST_AsGeoJSON(locations.geom)::JSONB AS geometry,
        devices.id AS device_id,
        devices.name AS device_name,
        sensors.id AS sensor_id,
        sensors.name AS sensor_name,
        MAX(location_history.created_at) AS lastest_timestamp,
        AVG(location_history.probability) AS average_probability,
        COUNT(measurements.*) AS scanners
    FROM
        location_history
    INNER JOIN devices
        ON devices.id = location_history.device_id
        ANd devices.username = 'admin'
        AND devices.is_deleted = false
    INNER JOIN locations
        ON locations.id = location_history.location_id
        AND locations.is_deleted = false

    INNER JOIN sensors
        ON sensors.device_id = devices.id
    LEFT JOIN measurements
        ON measurements.sensor_id = sensors.id
        AND measurements.created_at >= (NOW() - INTERVAL '60 minutes')

    WHERE
        location_history.created_at >= (NOW() - INTERVAL '60 minutes')
    GROUP BY locations.id, devices.id, sensors.id
)

SELECT json_agg(c)
FROM (
    SELECT
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry,
        json_agg(
            json_build_object(
                'device_id',
                device_id,
                'device_name',
                device_name,
                'lastest_timestamp',
                lastest_timestamp,
                'average_probability',
                average_probability,
                'sensor_id',
                sensor_id,
                'sensor_name',
                sensor_name,
                'scanners',
                scanners
            )
        ) AS devices
    FROM device_locations
    GROUP BY
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry
) c;




















-- GetByLocation

-- get count of mac addresses seen by device within time window

WITH device_sensor_locations AS (
    SELECT
        locations.id AS location_id,
        locations.name AS location_name,
        ST_AsGeoJSON(locations.geom)::JSONB AS geometry,
        devices.id AS device_id,
        devices.name AS device_name,
        sensors.id AS sensor_id,
        sensors.name AS sensor_name,
        MAX(location_history.created_at) AS lastest_timestamp,
        AVG(location_history.probability) AS average_probability,
        COUNT(measurements.*) AS scanners
    FROM
        location_history
    INNER JOIN devices
        ON devices.id = location_history.device_id
        ANd devices.username = 'admin'
        AND devices.is_deleted = false
    INNER JOIN locations
        ON locations.id = location_history.location_id
        AND locations.is_deleted = false

    INNER JOIN sensors
        ON sensors.device_id = devices.id
    LEFT JOIN measurements
        ON measurements.sensor_id = sensors.id
        AND measurements.created_at >= (NOW() - INTERVAL '60 minutes')

    WHERE
        location_history.created_at >= (NOW() - INTERVAL '60 minutes')
    GROUP BY locations.id, devices.id, sensors.id
),
device_locations AS (
    SELECT
        device_sensor_locations.location_id,
        device_sensor_locations.location_name,
        device_sensor_locations.geometry,
        device_sensor_locations.device_id,
        device_sensor_locations.device_name,
        device_sensor_locations.lastest_timestamp,
        device_sensor_locations.average_probability,
        json_agg(
            json_build_object(
                'sensor_id',
                device_sensor_locations.sensor_id,
                'sensor_name',
                device_sensor_locations.sensor_name,
                'scanners',
                device_sensor_locations.scanners
            )
        ) AS sensors
    FROM device_sensor_locations
    GROUP BY
        device_sensor_locations.location_id,
        device_sensor_locations.location_name,
        device_sensor_locations.geometry,
        device_sensor_locations.device_id,
        device_sensor_locations.device_name,
        device_sensor_locations.lastest_timestamp,
        device_sensor_locations.average_probability
)

SELECT json_agg(c)
FROM (
    SELECT
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry,
        json_agg(
            json_build_object(
                'device_id',
                device_locations.device_id,
                'device_name',
                device_locations.device_name,
                'lastest_timestamp',
                device_locations.lastest_timestamp,
                'average_probability',
                device_locations.average_probability,
                'sensors',
                device_locations.sensors
            )
        ) AS devices
    FROM device_locations
    GROUP BY
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry
) c;




























WITH
    device_sensor_locations AS (
        SELECT
            locations.id AS location_id,
            locations.name AS location_name,
            ST_AsGeoJSON(locations.geom)::JSONB AS geometry,
            devices.id AS device_id,
            devices.name AS device_name,
            sensors.id AS sensor_id,
            sensors.name AS sensor_name,
            MIN(location_history.created_at) AS first_timestamp,
            MAX(location_history.created_at) AS lastest_timestamp,
            AVG(location_history.probability) AS average_probability,
            COUNT(measurements.*) AS num_measurements
        FROM
            location_history
        INNER JOIN devices
            ON devices.id = location_history.device_id
            ANd devices.username = 'admin'
            AND devices.is_deleted = false
        INNER JOIN locations
            ON locations.id = location_history.location_id
            AND locations.is_deleted = false
        INNER JOIN sensors
            ON sensors.device_id = devices.id
        LEFT JOIN measurements
            ON measurements.sensor_id = sensors.id
            AND measurements.created_at >= (NOW() - INTERVAL '5 minutes')
        WHERE
            location_history.created_at >= (NOW() - INTERVAL '5 minutes')
        GROUP BY locations.id, devices.id, sensors.id
    ),
    device_locations AS (
        SELECT
            device_sensor_locations.location_id,
            device_sensor_locations.location_name,
            device_sensor_locations.geometry,
            device_sensor_locations.device_id,
            device_sensor_locations.device_name,
            device_sensor_locations.first_timestamp,
            device_sensor_locations.lastest_timestamp,
            device_sensor_locations.average_probability,
            json_agg(
                json_build_object(
                    'sensor_id',
                    device_sensor_locations.sensor_id,
                    'sensor_name',
                    device_sensor_locations.sensor_name,
                    'measurements',
                    device_sensor_locations.num_measurements
                )
            ) AS sensors
        FROM device_sensor_locations
        GROUP BY
            device_sensor_locations.location_id,
            device_sensor_locations.location_name,
            device_sensor_locations.geometry,
            device_sensor_locations.device_id,
            device_sensor_locations.device_name,
            device_sensor_locations.first_timestamp,
            device_sensor_locations.lastest_timestamp,
            device_sensor_locations.average_probability
    )

SELECT json_agg(c)
FROM (
    SELECT
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry,
        json_agg(
            json_build_object(
                'device_id',
                device_locations.device_id,
                'device_name',
                device_locations.device_name,
                'first_timestamp',
                device_locations.first_timestamp,
                'lastest_timestamp',
                device_locations.lastest_timestamp,
                'average_probability',
                device_locations.average_probability,
                'sensors',
                device_locations.sensors
            )
        ) AS devices
    FROM device_locations
    GROUP BY
        device_locations.location_id,
        device_locations.location_name,
        device_locations.geometry
) c;
