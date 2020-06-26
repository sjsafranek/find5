
DROP TABLE IF EXISTS users CASCADE;

-- @table users
-- @description stores users for find system
CREATE TABLE IF NOT EXISTS users (
    username        VARCHAR(50) NOT NULL PRIMARY KEY CHECK(username != ''),
    email           VARCHAR(50) NOT NULL CHECK(email != ''),
    apikey          VARCHAR(32) NOT NULL UNIQUE DEFAULT md5(random()::text),
    secret_token    VARCHAR(32) NOT NULL DEFAULT md5(random()::text),
    is_active       BOOLEAN DEFAULT TRUE,
    is_deleted      BOOLEAN DEFAULT FALSE,
    is_superuser    BOOLEAN DEFAULT FALSE,
    salt            VARCHAR DEFAULT gen_salt('bf', 8),
    password        VARCHAR,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

-- @function is_password
-- @description check user password
CREATE OR REPLACE FUNCTION is_password(usr TEXT, psw TEXT)
    RETURNS TEXT AS
$BODY$
BEGIN
    -- back door for using hashed password for login
    PERFORM * FROM users
        WHERE
            users.username = usr
        AND (
            users.password = psw
                OR
            users.password = crypt(psw, salt)
        );
    -- check results
    IF FOUND THEN
        RETURN TRUE;
    ELSE
        RETURN FALSE;
    END IF;
END;
$BODY$
LANGUAGE 'plpgsql';




-- This example cleans the input before it’s put into the database, in case someone accidentally put a space in their email address, or a line-break in their name.
-- Source: https://sivers.org/pg
-- TODO figure out line break issue...?
CREATE OR REPLACE FUNCTION clean_user()
RETURNS TRIGGER AS $$
    BEGIN
        NEW.username = btrim(regexp_replace(NEW.username, '\s+', ' ', 'g'));
        NEW.email = lower(regexp_replace(NEW.email, '\s', '', 'g'));
        RETURN NEW;
    END;
$$ LANGUAGE 'plpgsql';

DROP TRIGGER IF EXISTS users_clean ON users;
CREATE TRIGGER users_clean
    BEFORE INSERT OR UPDATE OF username, email ON users
        FOR EACH ROW EXECUTE PROCEDURE clean_user();



INSERT INTO users (username, email, password, is_superuser) VALUES('admin_user', 'admin_user', 'dev', TRUE);
