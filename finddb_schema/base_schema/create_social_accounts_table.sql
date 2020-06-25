
DROP TABLE IF EXISTS social_accounts CASCADE;
CREATE TABLE social_accounts (
	id		        VARCHAR,
	name	        VARCHAR,
	type	        VARCHAR,
	email			VARCHAR,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT account PRIMARY KEY(email, type),
	FOREIGN KEY (email) REFERENCES users(username) ON DELETE CASCADE
);

-- @trigger users_update
DROP TRIGGER IF EXISTS social_accounts_update ON social_accounts;
CREATE TRIGGER social_accounts_update
    BEFORE UPDATE ON social_accounts
        FOR EACH ROW
            EXECUTE PROCEDURE update_modified_column();
