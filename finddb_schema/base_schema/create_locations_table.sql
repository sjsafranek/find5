
DROP TABLE IF EXISTS locations CASCADE;

-- @table locations
-- @description stores location info for location_history and predictions
CREATE TABLE IF NOT EXISTS locations (
    id              VARCHAR(36) PRIMARY KEY DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR(50) NOT NULL,
    username        VARCHAR(50) NOT NULL,
    longitude       DOUBLE PRECISION,
    latitude        DOUBLE PRECISION,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE,
    UNIQUE(username, name)
);

-- add geometry column
SELECT AddGeometryColumn ('locations', 'geom', 4326, 'POINT', 2);
CREATE INDEX locations__gidx ON locations USING GIST(geom);

COMMENT ON TABLE locations IS 'location info';
COMMENT ON COLUMN locations.username IS 'the user that owns this location';
COMMENT ON COLUMN locations.name IS 'name of the location';
COMMENT ON COLUMN locations.geom IS 'POINT geometry of the location';

DROP TRIGGER IF EXISTS locations__update ON locations;
CREATE TRIGGER locations_update
    BEFORE UPDATE ON locations
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();

-- @function update_point_geom
-- @description updates geom column using longitude and latitude columns
CREATE OR REPLACE FUNCTION update_point_geom()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.geom = ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326);
        RETURN NEW;
    END;
$$ language 'plpgsql';

-- @trigger places_geom_update
-- @description updates geom in the places table
DROP TRIGGER IF EXISTS locations_geom_update ON locations;
CREATE TRIGGER locations_geom_update
    BEFORE INSERT OR UPDATE ON locations
        FOR EACH ROW
        WHEN (NEW.longitude IS NOT NULL AND NEW.latitude IS NOT NULL)
            EXECUTE PROCEDURE update_point_geom();
