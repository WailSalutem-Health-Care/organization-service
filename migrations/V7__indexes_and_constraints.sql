DO $$
DECLARE
    s RECORD;
BEGIN
    FOR s IN
        SELECT schema_name
        FROM wailsalutem.organizations
    LOOP
        EXECUTE format(
            'CREATE INDEX IF NOT EXISTS idx_%I_users_keycloak ON %I.users(keycloak_user_id)',
            s.schema_name, s.schema_name
        );
    END LOOP;
END $$;
