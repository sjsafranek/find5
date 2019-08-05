

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
