/*======================================================================*/
--  20191103.sql
--  :mode=pl-sql:tabSize=3:indentSize=3:
--  Fri Feb 15 15:58:26 PST 2019 @40 /Internet Time/
--  Purpose:
/*======================================================================*/

\set ON_ERROR_STOP on
set client_min_messages to 'warning';

-- \cd %%subdir%%

/*======================================================================*/
--  Prevent application of this patch; comment out when ready for use.
/*======================================================================*/

/*
do $$
begin
   raise exception 'This patch is not yet usable -- aborting';
end $$;
*/

/*======================================================================*/
--  Test the database version.
/*======================================================================*/
do $$
DECLARE
   vers       text;
   check_vers text:='5.0.1';
BEGIN
   SELECT value INTO vers FROM config WHERE key='version';
   IF vers!=check_vers THEN
      raise exception 'Version % was not expected -- aborting',vers
         USING hint = ' Expected version is '||check_vers;
	END IF;
END $$;

/*======================================================================*/
-- Sub-scripts.
/*======================================================================*/
-- these are found in expanded_schema dir:
-- \i 20190730_additional.sql

-- ALTER TABLE users ADD COLUMN is_superuser BOOLEAN;
-- INSERT INTO users (username, email, password, is_superuser) VALUES('admin_user', 'admin_user', 'dev', TRUE);
--
-- DROP VIEW IF EXISTS users_view;
-- CREATE OR REPLACE VIEW users_view AS (
--     SELECT
--         *,
--         json_build_object(
--             'email', email,
--             'username', username,
--             'apikey', apikey,
--             'secret_token', secret_token,
--             'is_active', is_active,
--             'is_deleted', is_deleted,
--             'is_superuser', is_superuser,
--             'created_at', to_char(created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
--             'updated_at', to_char(updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"')
--         ) AS user_json
--     FROM users
-- );



/*======================================================================*/
--  Rev database version.
/*======================================================================*/
UPDATE
   config
SET
   value='5.0.2'
WHERE
   key='version';
