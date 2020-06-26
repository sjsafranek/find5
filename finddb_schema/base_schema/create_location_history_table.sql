
DROP TABLE IF EXISTS location_history CASCADE;

-- @table location_history
-- @description stores device location history
CREATE TABLE IF NOT EXISTS location_history (
    id                      SERIAL PRIMARY KEY,
    device_id               VARCHAR(36) NOT NULL,
    location_id             VARCHAR(36) NOT NULL,
    probability             REAL,
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    unique_ts               INTEGER DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP)::INTEGER,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE,
    CONSTRAINT unique_device_location UNIQUE(device_id, location_id, probability, unique_ts)
);

COMMENT ON TABLE location_history IS 'location history of devices';
COMMENT ON COLUMN location_history.device_id IS 'the device that is at the location';
COMMENT ON COLUMN location_history.location_id IS 'location of the device';
COMMENT ON COLUMN location_history.probability IS 'probability of the device being at the location';
