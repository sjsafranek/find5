package database

// TABLES_SQL defines the main database tables
// and trigger functions.
var TABLES_SQL = `
    CREATE TABLE IF NOT EXISTS keystore (
        key TEXT NOT NULL PRIMARY KEY,
        value TEXT,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS keystore_key_idx ON keystore(key);
    CREATE TRIGGER IF NOT EXISTS keystore__update
        AFTER
        UPDATE
        ON keystore
        FOR EACH ROW
    BEGIN
        UPDATE keystore SET update_at=CURRENT_TIMESTAMP WHERE timestamp=OLD.timestamp;
    END;


    CREATE TABLE IF NOT EXISTS sensors (
        timestamp INTEGER NOT NULL PRIMARY KEY,
        deviceid INTEGER,
        locationid INTEGER,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        -- wifi TEXT,
        sensor_type TEXT,
        sensor TEXT,
        UNIQUE(timestamp)
    );
    CREATE INDEX IF NOT EXISTS sensors_devices ON sensors (deviceid);
    CREATE TRIGGER IF NOT EXISTS sensors__update
        AFTER
        UPDATE
        ON sensors
        FOR EACH ROW
    BEGIN
        UPDATE sensors SET update_at=CURRENT_TIMESTAMP WHERE timestamp=OLD.timestamp;
    END;


    CREATE TABLE IF NOT EXISTS location_predictions (
        timestamp INTEGER NOT NULL PRIMARY KEY,
        locationid TEXT,
        probability TEXT,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(timestamp)
    );
    CREATE TRIGGER IF NOT EXISTS location_predictions__update
        AFTER
        UPDATE
        ON location_predictions
        FOR EACH ROW
    BEGIN
        UPDATE location_predictions SET update_at=CURRENT_TIMESTAMP WHERE timestamp=OLD.timestamp;
    END;


    CREATE TABLE IF NOT EXISTS gps (
        id INTEGER PRIMARY KEY,
        mac TEXT,
        loc TEXT,
        lat REAL,
        lon REAL,
        alt REAL,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TRIGGER IF NOT EXISTS gps__update
        AFTER
        UPDATE
        ON gps
        FOR EACH ROW
    BEGIN
        UPDATE gps SET update_at=CURRENT_TIMESTAMP WHERE timestamp=OLD.timestamp;
    END;


    CREATE TABLE IF NOT EXISTS calibrations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        probability_means TEXT,
        probabilities_of_best_guess TEXT,
        percent_correct TEXT,
        accuracy_breakdown TEXT,
        prediction_analysis TEXT,
        algorithm_efficacy TEXT,
        calibration_time DATETIME DEFAULT CURRENT_TIMESTAMP,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );


    CREATE TABLE IF NOT EXISTS users (
        username TEXT,
        password TEXT,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

`

// CREATE TABLE IF NOT EXISTS learning (
//     id INTEGER PRIMARY KEY AUTOINCREMENT,
//     algorithm TEXT,
//     data TEXT,
//     create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//     update_at DATETIME DEFAULT CURRENT_TIMESTAMP
// );
