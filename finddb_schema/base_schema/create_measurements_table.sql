
DROP TABLE IF EXISTS measurements CASCADE;

-- @table measurements
-- @description stores measurements collected by sensors at a given location
CREATE TABLE IF NOT EXISTS measurements (
    id              SERIAL PRIMARY KEY,
    location_id     VARCHAR(36) REFERENCES locations(id) ON DELETE CASCADE,
    sensor_id       VARCHAR(36),
    key             VARCHAR(50),
    value           DOUBLE PRECISION,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE CASCADE,
    -- TODO: should be down in a database patch
    UNIQUE(sensor_id, value, key, created_at)
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
            (SELECT
                devices.id,
                NEW.location_id,
                100.0
            FROM sensors
                INNER JOIN devices
                    ON devices.id = sensors.device_id
                WHERE
                    sensors.id = NEW.sensor_id
            )
        ON CONFLICT DO NOTHING;
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
