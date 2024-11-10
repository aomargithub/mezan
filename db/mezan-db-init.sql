--as mentioned in https://wiki.postgresql.org/wiki/Database_Schema_Recommendations_for_an_Application, public schema should be deleted
DROP SCHEMA IF EXISTS public cascade;
CREATE SCHEMA base;
CREATE SCHEMA mezan;