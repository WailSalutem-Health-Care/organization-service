CREATE OR REPLACE FUNCTION wailsalutem.create_tenant_schema(schema_name TEXT)
RETURNS void AS $$
BEGIN
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            keycloak_user_id UUID NOT NULL,
            employee_id VARCHAR(50) UNIQUE,
            first_name VARCHAR(100),
            last_name VARCHAR(100),
            email VARCHAR(255),
            phone_number VARCHAR(50),
            role VARCHAR(50),
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.patients (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            patient_id VARCHAR(50) UNIQUE,
            first_name VARCHAR(100),
            last_name VARCHAR(100),
            keycloak_user_id UUID,
            email VARCHAR(255),
            phone_number VARCHAR(50),
            date_of_birth DATE,
            address TEXT,
            emergency_contact_name VARCHAR(255),
            emergency_contact_phone VARCHAR(50),
            medical_notes TEXT,
            careplan_type VARCHAR(100),
            careplan_frequency VARCHAR(100),
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )', schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.care_sessions (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            session_id VARCHAR(50) UNIQUE,
            patient_id UUID REFERENCES %I.patients(id),
            caregiver_id UUID,
            check_in_time TIMESTAMP,
            check_out_time TIMESTAMP,
            status VARCHAR(50),
            caregiver_notes TEXT,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )', schema_name, schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.nfc_tags (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            tag_id VARCHAR(100) UNIQUE,
            patient_id UUID REFERENCES %I.patients(id),
            issued_at TIMESTAMP DEFAULT now(),
            status VARCHAR(50),
            deactivated_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP
        )', schema_name, schema_name);

    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I.feedback (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            care_session_id UUID UNIQUE,
            patient_id UUID,
            caregiver_id UUID,
            rating INTEGER CHECK (rating BETWEEN 1 AND 5),
            patient_feedback TEXT,
            created_at TIMESTAMP DEFAULT now(),
            deleted_at TIMESTAMP
        )', schema_name);
END;
$$ LANGUAGE plpgsql;
