

DROP TABLE IF EXISTS sensors CASCADE;

-- @table sensors
-- @description stores device sensor metadata
CREATE TABLE IF NOT EXISTS sensors (
    id              VARCHAR(36) PRIMARY KEY DEFAULT md5(random()::text || now()::text)::uuid,
    device_id       VARCHAR(36) NOT NULL CHECK(device_id != ''),
    name            VARCHAR(50) NOT NULL CHECK(name != ''),
    type            VARCHAR(50) DEFAULT 'unknown',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    is_active       BOOLEAN DEFAULT true,
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    UNIQUE(device_id, name)
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
