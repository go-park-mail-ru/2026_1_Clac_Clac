-- init_service_users.sql
-- Создает сервисную учетную запись с минимально необходимыми правами
-- для runtime-работы приложения.
--
-- Параметры (передаются через psql -v):
--   service_user      — имя сервисного пользователя (обязательный)
--   service_password  — пароль сервисного пользователя (обязательный)
--   admin_user        — имя владельца БД для ALTER DEFAULT PRIVILEGES (обязательный)
--   permissions       — набор прав на таблицы (по умолчанию 'SELECT, INSERT, UPDATE, DELETE')
--
-- Примеры:
--
--   # Board (полный набор прав):
--   psql -v service_user="'board_service'" \
--        -v service_password="'secure_password'" \
--        -v admin_user="'board_admin'" \
--        -f init_service_users.sql
--
--   # User (без DELETE — профили не удаляются):
--   psql -v service_user="'user_service'" \
--        -v service_password="'secure_password'" \
--        -v admin_user="'user_admin'" \
--        -v permissions="'SELECT, INSERT, UPDATE'" \
--        -f init_service_users.sql

\if :{?permissions}
\else
    \set permissions 'SELECT, INSERT, UPDATE, DELETE'
\endif

SELECT format('CREATE ROLE %I WITH LOGIN PASSWORD %L', :'service_user', :'service_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'service_user')
\gexec

SELECT format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), :'service_user') \gexec
GRANT USAGE ON SCHEMA public TO :"service_user";

SELECT format('GRANT %s ON ALL TABLES IN SCHEMA public TO %I', :'permissions', :'service_user') \gexec
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO :"service_user";


SELECT format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT %s ON TABLES TO %I',
    :'admin_user', :'permissions', :'service_user'
) \gexec
SELECT format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %I',
    :'admin_user', :'service_user'
) \gexec
