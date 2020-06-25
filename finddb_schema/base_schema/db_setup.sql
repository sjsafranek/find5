/*======================================================================*/
--  db_setup.sql
--   -- :mode=pl-sql:tabSize=3:indentSize=3:
--  Mon Aug 17 14:44:44 PST 2015 @144 /Internet Time/
--  Purpose:
--  NOTE: must be connected as 'postgres' user or a superuser to start.
/*======================================================================*/

\set ON_ERROR_STOP on
set client_min_messages to 'warning';


\i create_extensions.sql
\i create_general_functions.sql
\i create_config_table.sql

\i create_users_table.sql
\i create_social_accounts_table.sql
\i create_users_view.sql

\i create_devices_table.sql
\i create_locations_table.sql
\i create_location_history_table.sql
\i create_sensors_table.sql
\i create_measurements_table.sql
