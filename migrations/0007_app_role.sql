-- 0007_app_role.sql
-- The least-privilege application role. internal/tenancy does `SET LOCAL ROLE
-- netctl_app` at the start of every tenant-scoped transaction, so RLS is enforced
-- for tenant operations REGARDLESS of how the control plane authenticated to
-- Postgres (even a superuser session is filtered once it assumes this role).
-- netctl_app is NOSUPERUSER NOBYPASSRLS so it can never bypass row security.
--
-- The login/migration role must be able to assume netctl_app: a superuser always
-- can; otherwise grant membership (`GRANT netctl_app TO <login_role>`).

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'netctl_app') THEN
        CREATE ROLE netctl_app NOLOGIN NOSUPERUSER NOBYPASSRLS;
    END IF;
END $$;

GRANT USAGE ON SCHEMA public TO netctl_app;

GRANT SELECT, INSERT, UPDATE, DELETE ON
    organizations, teams, projects, users, service_accounts,
    roles, role_permissions, role_bindings,
    agents, tests, results
TO netctl_app;

-- Read-only catalog.
GRANT SELECT ON permissions TO netctl_app;

-- Audit is append-only for the application role (no UPDATE/DELETE).
GRANT SELECT, INSERT ON audit_events TO netctl_app;

-- Future tenant-owned tables created by the migration role inherit DML grants so
-- later sprints' tables are reachable by netctl_app without re-granting.
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO netctl_app;
