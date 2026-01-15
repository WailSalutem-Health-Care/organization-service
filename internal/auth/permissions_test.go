package auth

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadPermissions_Success tests successfully loading permissions from YAML
func TestLoadPermissions_Success(t *testing.T) {
	// Create a temporary permissions file
	tmpDir := t.TempDir()
	permFile := filepath.Join(tmpDir, "permissions.yml")

	content := `roles:
  SUPER_ADMIN:
    - organization:create
    - organization:view
    - organization:delete
  ORG_ADMIN:
    - organization:view
    - user:create
  PATIENT:
    - patient:view
    - patient:update
`

	err := os.WriteFile(permFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test permissions file: %v", err)
	}

	// Load permissions
	perms, err := LoadPermissions(permFile)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if perms == nil {
		t.Fatal("Expected permissions map, got nil")
	}

	// Check SUPER_ADMIN permissions
	superAdminPerms, exists := perms["SUPER_ADMIN"]
	if !exists {
		t.Error("Expected SUPER_ADMIN role to exist")
	}
	if len(superAdminPerms) != 3 {
		t.Errorf("Expected 3 permissions for SUPER_ADMIN, got %d", len(superAdminPerms))
	}
	if !contains(superAdminPerms, "organization:create") {
		t.Error("Expected SUPER_ADMIN to have 'organization:create' permission")
	}

	// Check ORG_ADMIN permissions
	orgAdminPerms, exists := perms["ORG_ADMIN"]
	if !exists {
		t.Error("Expected ORG_ADMIN role to exist")
	}
	if len(orgAdminPerms) != 2 {
		t.Errorf("Expected 2 permissions for ORG_ADMIN, got %d", len(orgAdminPerms))
	}

	// Check PATIENT permissions
	patientPerms, exists := perms["PATIENT"]
	if !exists {
		t.Error("Expected PATIENT role to exist")
	}
	if len(patientPerms) != 2 {
		t.Errorf("Expected 2 permissions for PATIENT, got %d", len(patientPerms))
	}
}

// TestLoadPermissions_FileNotFound tests loading non-existent file
func TestLoadPermissions_FileNotFound(t *testing.T) {
	perms, err := LoadPermissions("/nonexistent/path/permissions.yml")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
	if perms != nil {
		t.Error("Expected nil permissions, got non-nil")
	}
}

// TestLoadPermissions_InvalidYAML tests loading invalid YAML
func TestLoadPermissions_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	permFile := filepath.Join(tmpDir, "bad_permissions.yml")

	// Write invalid YAML
	content := `roles:
  SUPER_ADMIN:
    - organization:create
    invalid yaml structure here
      - no proper indentation
`

	err := os.WriteFile(permFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	perms, err := LoadPermissions(permFile)

	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
	if perms != nil {
		t.Error("Expected nil permissions for invalid YAML")
	}
}

// TestLoadPermissions_EmptyFile tests loading empty permissions file
func TestLoadPermissions_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	permFile := filepath.Join(tmpDir, "empty_permissions.yml")

	// Write empty file
	err := os.WriteFile(permFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	perms, err := LoadPermissions(permFile)

	// Should succeed with nil or empty map (both are acceptable)
	if err != nil {
		t.Errorf("Expected no error for empty file, got: %v", err)
	}
	// Empty file results in nil map which is acceptable
	if perms != nil && len(perms) != 0 {
		t.Errorf("Expected 0 roles, got %d", len(perms))
	}
}

// TestLoadPermissions_EmptyRoles tests file with roles but no permissions
func TestLoadPermissions_EmptyRoles(t *testing.T) {
	tmpDir := t.TempDir()
	permFile := filepath.Join(tmpDir, "empty_roles.yml")

	content := `roles:
  SUPER_ADMIN: []
  ORG_ADMIN: []
`

	err := os.WriteFile(permFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	perms, err := LoadPermissions(permFile)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	superAdminPerms, exists := perms["SUPER_ADMIN"]
	if !exists {
		t.Error("Expected SUPER_ADMIN role to exist")
	}
	if len(superAdminPerms) != 0 {
		t.Errorf("Expected 0 permissions for SUPER_ADMIN, got %d", len(superAdminPerms))
	}
}

// TestLoadPermissions_RealFile tests loading the actual permissions.yml
func TestLoadPermissions_RealFile(t *testing.T) {
	// This test assumes permissions.yml exists in the project root
	// Skip if running in isolation
	permFile := "../../permissions.yml"
	
	if _, err := os.Stat(permFile); os.IsNotExist(err) {
		t.Skip("Skipping test: permissions.yml not found (expected when running isolated tests)")
	}

	perms, err := LoadPermissions(permFile)

	if err != nil {
		t.Fatalf("Expected to load real permissions.yml, got error: %v", err)
	}
	if perms == nil {
		t.Fatal("Expected permissions map, got nil")
	}

	// Verify expected roles exist
	expectedRoles := []string{"SUPER_ADMIN", "ORG_ADMIN", "CAREGIVER", "PATIENT", "MUNICIPALITY", "INSURER"}
	for _, role := range expectedRoles {
		if _, exists := perms[role]; !exists {
			t.Errorf("Expected role '%s' to exist in permissions.yml", role)
		}
	}

	// Verify SUPER_ADMIN has comprehensive permissions
	superAdminPerms := perms["SUPER_ADMIN"]
	expectedPerms := []string{
		"organization:create",
		"organization:view",
		"organization:update",
		"organization:delete",
		"user:create",
		"user:view",
		"user:update",
		"user:delete",
	}
	for _, perm := range expectedPerms {
		if !contains(superAdminPerms, perm) {
			t.Errorf("Expected SUPER_ADMIN to have permission '%s'", perm)
		}
	}

	// Verify ORG_ADMIN has limited permissions
	orgAdminPerms := perms["ORG_ADMIN"]
	if contains(orgAdminPerms, "organization:create") {
		t.Error("ORG_ADMIN should not have 'organization:create' permission")
	}
	if contains(orgAdminPerms, "organization:delete") {
		t.Error("ORG_ADMIN should not have 'organization:delete' permission")
	}
}

// Helper function to check if slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
