
DROP VIEW IF EXISTS locations_view CASCADE;

-- CREATE OR REPLACE VIEW locations_view AS (
--     SELECT
--         *,
--         json_build_object(
--             'type', 'Feature',
--             'geometry', ST_AsGeoJSON(geom)::jsonb,
--             'properties', json_build_object(
--                 'id', id,
--                 'name', name,
--                 -- 'username', username,
--                 'created_at', to_char(locations.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
--                 'updated_at', to_char(locations.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"')
--             )
--         ) AS location_json
--     FROM locations
-- );
