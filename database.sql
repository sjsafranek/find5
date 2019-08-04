-- Enable pgcrypto for passwords
CREATE EXTENSION pgcrypto;

-- Enable PostGIS (includes raster)
CREATE EXTENSION postgis;
-- Enable Topology
CREATE EXTENSION postgis_topology;
-- Enable PostGIS Advanced 3D
-- and other geoprocessing algorithms
-- sfcgal not available with all distributions
-- CREATE EXTENSION postgis_sfcgal;
-- fuzzy matching needed for Tiger
-- CREATE EXTENSION fuzzystrmatch;
-- rule based standardizer
-- CREATE EXTENSION address_standardizer;
-- example rule data set
-- CREATE EXTENSION address_standardizer_data_us;
-- Enable US Tiger Geocoder
-- CREATE EXTENSION postgis_tiger_geocoder;



CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = now();
        RETURN NEW;
    END;
$$ language 'plpgsql';



CREATE TABLE IF NOT EXISTS users (
    email           VARCHAR,
    username        VARCHAR NOT NULL UNIQUE,
    apikey          VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text),
    secret_token    VARCHAR NOT NULL DEFAULT md5(random()::text),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    salt            VARCHAR DEFAULT gen_salt('bf', 8),
    password        VARCHAR,
    PRIMARY KEY(username)
);
COMMENT ON TABLE users IS 'username and password info';

 -- update triggers
DROP TRIGGER IF EXISTS users_update ON users;
CREATE TRIGGER users_update
    BEFORE UPDATE ON users
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();
 -- .end

 -- hash user password
CREATE OR REPLACE FUNCTION hash_password()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.password = crypt(NEW.password, NEW.salt);
        RETURN NEW;
    END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS users_password_insert ON users;
DROP TRIGGER IF EXISTS users_password_update ON users;
-- hash password on insert
CREATE TRIGGER users_password_insert
    BEFORE INSERT ON users
        FOR EACH ROW
            EXECUTE PROCEDURE hash_password();
-- check if password changed
CREATE TRIGGER users_password_update
    BEFORE UPDATE ON users
        FOR EACH ROW
        WHEN (OLD.password IS DISTINCT FROM NEW.password)
            EXECUTE PROCEDURE hash_password();



CREATE TABLE IF NOT EXISTS devices (
    id              VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR(50),
    type            VARCHAR(50),
    username        VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);
COMMENT ON TABLE devices IS 'device info for data collection';

DROP TRIGGER IF EXISTS devices_update ON devices;
CREATE TRIGGER device_update
    BEFORE UPDATE ON devices
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();



CREATE TABLE IF NOT EXISTS locations (
    id              VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR,
    username        VARCHAR,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);
COMMENT ON TABLE locations IS 'location info';

SELECT AddGeometryColumn ('locations', 'geom', 4326, 'POINT', 2);
CREATE INDEX locations__gidx ON locations USING GIST(geom);

DROP TRIGGER IF EXISTS locations__update ON locations;
CREATE TRIGGER locations_update
    BEFORE UPDATE ON locations
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();



CREATE TABLE IF NOT EXISTS location_history (
    id                      SERIAL PRIMARY KEY,
    device_id               VARCHAR,
    location_id             VARCHAR,
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);
COMMENT ON TABLE location_history IS 'stores location history of devices';



CREATE TABLE IF NOT EXISTS sensors (
    id              VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    device_id       VARCHAR,
    name            VARCHAR,
    type            VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);
COMMENT ON TABLE sensors IS 'stores device sensor info';

DROP TRIGGER IF EXISTS sensors_update ON sensors;
CREATE TRIGGER sensors_update
    BEFORE UPDATE ON sensors
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();




CREATE TABLE IF NOT EXISTS measurements (
    id SERIAL PRIMARY KEY,
    -- event_timestamp INTEGER,
    -- id              PRIMARY KEY SERIAL AUTOINCRIMENT,
    location_id     VARCHAR,
    sensor_id       VARCHAR,
    key             VARCHAR,
    value           DOUBLE PRECISION,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);
COMMENT ON TABLE measurements IS 'stores measurements collected by device sensors at a given location';




-- COMMENT ON COLUMN inrix_xdvolumes.XD_ID IS 'road segment id';




-- DROP TRIGGER IF EXISTS measurements_update ON measurements;
-- CREATE TRIGGER measurements_update
--     BEFORE UPDATE ON measurements
--         FOR EACH ROW
--             EXECUTE PROCEDURE update_modified_column();
