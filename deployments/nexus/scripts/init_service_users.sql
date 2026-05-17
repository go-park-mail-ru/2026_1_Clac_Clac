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

-- Значение по умолчанию для permissions
\if :{?permissions}
\else
    \set permissions '\'SELECT, INSERT, UPDATE, DELETE\''
\endif

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :service_user) THEN
        EXECUTE format('CREATE USER %I WITH PASSWORD %L', :service_user, :service_password);
        RAISE NOTICE 'Created user: %', :service_user;
    ELSE
        RAISE NOTICE 'User already exists: %', :service_user;
    END IF;
END
$$;

GRANT CONNECT ON DATABASE current_database() TO :service_user;
GRANT USAGE ON SCHEMA public TO :service_user;

EXECUTE format('GRANT %s ON ALL TABLES IN SCHEMA public TO %I', :permissions, :service_user);
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO :service_user;

EXECUTE format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT %s ON TABLES TO %I',
    :admin_user, :permissions, :service_user
);
EXECUTE format(
    'ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %I',
    :admin_user, :service_user
);

RAISE NOTICE 'Service user % configured with permissions: %', :service_user, :permissions;
