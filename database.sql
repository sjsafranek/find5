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



-- TODO
-- add constrains for "name" columns
-- devices and sensors tables
-- locations table


CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = now();
        RETURN NEW;
    END;
$$ language 'plpgsql';


-- {USERS}
-- @table users
-- @description stores users for find system
CREATE TABLE IF NOT EXISTS users (
    email           VARCHAR(50),
    username        VARCHAR(50) NOT NULL UNIQUE,
    apikey          VARCHAR(32) NOT NULL UNIQUE DEFAULT md5(random()::text),
    secret_token    VARCHAR(32) NOT NULL DEFAULT md5(random()::text),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    salt            VARCHAR DEFAULT gen_salt('bf', 8),
    password        VARCHAR,
    PRIMARY KEY(username)
);

COMMENT ON TABLE users IS 'username and password info';

-- @trigger users_update
DROP TRIGGER IF EXISTS users_update ON users;
CREATE TRIGGER users_update
    BEFORE UPDATE ON users
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();

 -- @method hash_password
 -- @description hashes password before storing in row
CREATE OR REPLACE FUNCTION hash_password()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.password = crypt(NEW.password, NEW.salt);
        RETURN NEW;
    END;
$$ language 'plpgsql';

-- @trigger users_password_insert
-- @description trigger hash password
DROP TRIGGER IF EXISTS users_password_insert ON users;
CREATE TRIGGER users_password_insert
    BEFORE INSERT ON users
        FOR EACH ROW
            EXECUTE PROCEDURE hash_password();

-- @trigger users_password_update
-- @description trigger hash password if password has changed
DROP TRIGGER IF EXISTS users_password_update ON users;
CREATE TRIGGER users_password_update
    BEFORE UPDATE ON users
        FOR EACH ROW
        WHEN (OLD.password IS DISTINCT FROM NEW.password)
            EXECUTE PROCEDURE hash_password();

-- END {USERS}



-- {USER DEVICES}
-- @table devices
-- @description stores users devices
CREATE TABLE IF NOT EXISTS devices (
    id              VARCHAR(36) NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR(50),
    type            VARCHAR(50),
    username        VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);

COMMENT ON TABLE devices IS 'device info for data collection';
COMMENT ON COLUMN devices.username IS 'the user that owns the device';
COMMENT ON COLUMN devices.type IS 'the type of device (i.e. computer, phone, etc.)';
COMMENT ON COLUMN devices.name IS 'name of the device';

-- @trigger devices_update
DROP TRIGGER IF EXISTS devices_update ON devices;
CREATE TRIGGER device_update
    BEFORE UPDATE ON devices
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();

-- END {USER DEVICES}



-- {LOCATIONS}
-- @table locations
-- @description stores location info for location_history and predictions
CREATE TABLE IF NOT EXISTS locations (
    id              VARCHAR(36) NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR(50),
    username        VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
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

-- END {LOCATIONS}



-- {DEVICE LOCATION HISTORY}
-- @table location_history
-- @description stores device location history
CREATE TABLE IF NOT EXISTS location_history (
    id                      SERIAL PRIMARY KEY,
    device_id               VARCHAR(36),
    location_id             VARCHAR(36),
    probability             REAL,
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

COMMENT ON TABLE location_history IS 'location history of devices';
COMMENT ON COLUMN location_history.device_id IS 'the device that is at the location';
COMMENT ON COLUMN location_history.location_id IS 'location of the device';
COMMENT ON COLUMN location_history.probability IS 'probability of the device being at the location';

-- END {DEVICE LOCATION HISTORY}



-- {DEVICE SENSORS}
-- @table sensors
-- @description stores device sensor metadata
CREATE TABLE IF NOT EXISTS sensors (
    id              VARCHAR(36) NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    device_id       VARCHAR(36),
    name            VARCHAR(50),
    type            VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

COMMENT ON TABLE sensors IS 'device sensor info';
COMMENT ON COLUMN sensors.device_id IS 'device that the sensor belongs to';
COMMENT ON COLUMN sensors.type IS 'the type of sensor';

-- @trigger sensors_update
DROP TRIGGER IF EXISTS sensors_update ON sensors;
CREATE TRIGGER sensors_update
    BEFORE UPDATE ON sensors
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();

-- END {DEVICE SENSORS}



-- {SENSOR MEASUREMENTS}
-- @table measurements
-- @description stores measurements collected by sensors at a given location
CREATE TABLE IF NOT EXISTS measurements (
    id SERIAL PRIMARY KEY,
    location_id     VARCHAR(36) REFERENCES locations(id),
    sensor_id       VARCHAR(36),
    key             VARCHAR(50),
    value           DOUBLE PRECISION,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE
);

COMMENT ON TABLE measurements IS 'measurements collected by device sensors at a given location';
COMMENT ON COLUMN measurements.sensor_id IS 'sensor that created the measurement';
COMMENT ON COLUMN measurements.location_id IS 'location where the measurement was made';

-- @method measurements__location_history__insert
-- @description stores location_history record for sensor's device
CREATE OR REPLACE FUNCTION measurements__location_history__insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.location_id IS NOT NULL THEN
        INSERT INTO location_history (device_id, location_id, probability)
            SELECT
                devices.id,
                NEW.location_id,
                100.0
            FROM sensors
                INNER JOIN devices
                    ON devices.id = sensors.device_id
                WHERE
                    sensors.id = NEW.sensor_id
                ;
    END IF;
	RETURN NEW;
END;
$$ language 'plpgsql';

-- @trigger measurements__location_history__insert
DROP TRIGGER IF EXISTS measurements__location_history__insert ON measurements;
CREATE TRIGGER measurements__location_history__insert
    AFTER INSERT ON measurements
        FOR EACH ROW
            EXECUTE PROCEDURE measurements__location_history__insert();

-- END {SENSOR MEASUREMENTS}
