
DROP TABLE IF EXISTS devices CASCADE;

-- @table devices
-- @description stores users devices
CREATE TABLE IF NOT EXISTS devices (
    id              VARCHAR(36) PRIMARY KEY DEFAULT md5(random()::text || now()::text)::uuid,
    name            VARCHAR(50),
    type            VARCHAR(50),
    username        VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_deleted      BOOLEAN DEFAULT false,
    is_active       BOOLEAN DEFAULT true,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE,
    UNIQUE(username, name)
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
