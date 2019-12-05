
DROP TABLE IF EXISTS config CASCADE;

-- @table config
-- @description stores database config info
CREATE TABLE IF NOT EXISTS config (
    key             VARCHAR(50) PRIMARY KEY,
    value           VARCHAR(50),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE config IS 'Configuration and system state (key:value)';

-- @trigger config_update
-- @description update config record
DROP TRIGGER IF EXISTS config_update ON config;
CREATE TRIGGER config_update
    BEFORE UPDATE ON config
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();

INSERT INTO config(key, value) VALUES('version', '5.0.1');
