DO $$
DECLARE
    s RECORD;
BEGIN
    FOR s IN
        SELECT schema_name
        FROM wailsalutem.organizations
    LOOP
        PERFORM wailsalutem.create_tenant_schema(s.schema_name);
    END LOOP;
END $$;
