// go:build integration
//go:build integration

package users

import (
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
	"github.com/google/uuid"
)

// TestRepositoryCreate_Integration tests creating a user in tenant schema
func TestRepositoryCreate_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	// Create test organization and schema
	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_a")

	repo := NewRepository(db, nil)

	user := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "test@example.com",
		FirstName:      "Test",
		LastName:       "User",
		PhoneNumber:    "+1234567890",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == "" {
		t.Error("Expected user ID to be set")
	}

	if user.EmployeeID == "" {
		t.Error("Expected employee ID to be generated")
	}

	if user.EmployeeID != "EMP-0001" {
		t.Errorf("Expected first employee ID to be EMP-0001, got %s", user.EmployeeID)
	}

	if !user.IsActive {
		t.Error("Expected user to be active by default")
	}
}

// TestRepositoryGetByID_Integration tests retrieving a user
func TestRepositoryGetByID_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_b")
	repo := NewRepository(db, nil)

	// Create user
	created := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "get@test.com",
		FirstName:      "Get",
		LastName:       "User",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(created)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get user
	user, err := repo.GetByID(schemaName, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, user.ID)
	}

	if user.Email != created.Email {
		t.Errorf("Expected email %s, got %s", created.Email, user.Email)
	}

	if user.EmployeeID != "EMP-0001" {
		t.Errorf("Expected employee ID EMP-0001, got %s", user.EmployeeID)
	}
}

// TestRepositoryGetByID_CrossTenantIsolation_Integration tests tenant isolation
func TestRepositoryGetByID_CrossTenantIsolation_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	// Create two organizations
	org1ID, schema1 := testutil.CreateTestOrg(t, db, "hospital_1")
	org2ID, schema2 := testutil.CreateTestOrg(t, db, "hospital_2")

	repo := NewRepository(db, nil)

	// Create user in org1
	user1 := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "user1@org1.com",
		FirstName:      "User",
		LastName:       "One",
		Role:           "CAREGIVER",
		OrgID:          org1ID,
		OrgSchemaName:  schema1,
	}

	err := repo.Create(user1)
	if err != nil {
		t.Fatalf("Create user in org1 failed: %v", err)
	}

	// Try to get user1 from org2 schema (should fail)
	_, err = repo.GetByID(schema2, user1.ID)
	if err == nil {
		t.Error("Expected error when accessing user from different tenant, got nil")
	}

	// Create user in org2
	user2 := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "user2@org2.com",
		FirstName:      "User",
		LastName:       "Two",
		Role:           "CAREGIVER",
		OrgID:          org2ID,
		OrgSchemaName:  schema2,
	}

	err = repo.Create(user2)
	if err != nil {
		t.Fatalf("Create user in org2 failed: %v", err)
	}

	// Verify user2 is accessible from org2
	retrieved, err := repo.GetByID(schema2, user2.ID)
	if err != nil {
		t.Fatalf("GetByID from own tenant failed: %v", err)
	}

	if retrieved.ID != user2.ID {
		t.Errorf("Expected user ID %s, got %s", user2.ID, retrieved.ID)
	}

	// Verify user1 is still only in org1
	retrieved, err = repo.GetByID(schema1, user1.ID)
	if err != nil {
		t.Fatalf("GetByID from own tenant failed: %v", err)
	}

	if retrieved.ID != user1.ID {
		t.Errorf("Expected user ID %s, got %s", user1.ID, retrieved.ID)
	}
}

// TestRepositoryList_Integration tests listing users
func TestRepositoryList_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_c")
	repo := NewRepository(db, nil)

	// Create multiple users
	for i := 1; i <= 3; i++ {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          "list" + string(rune('0'+i)) + "@test.com",
			FirstName:      "List",
			LastName:       string(rune('A' + i - 1)),
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
	}

	// List users
	users, err := repo.List(schemaName)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(users) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(users))
	}
}

// TestRepositoryListWithPagination_Integration tests paginated user listing
func TestRepositoryListWithPagination_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_d")
	repo := NewRepository(db, nil)

	// Create 5 users
	for i := 1; i <= 5; i++ {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          "page" + string(rune('0'+i)) + "@test.com",
			FirstName:      "Page",
			LastName:       string(rune('A' + i - 1)),
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
	}

	// Get first page (limit 2)
	users, total, err := repo.ListWithPagination(schemaName, 2, 0, "")
	if err != nil {
		t.Fatalf("ListWithPagination failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users on first page, got %d", len(users))
	}

	if total < 5 {
		t.Errorf("Expected total >= 5, got %d", total)
	}

	// Get second page
	users, _, err = repo.ListWithPagination(schemaName, 2, 2, "")
	if err != nil {
		t.Fatalf("ListWithPagination page 2 failed: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users on second page, got %d", len(users))
	}
}

// TestRepositoryListActiveUsersByRole_Integration tests listing active users by role
func TestRepositoryListActiveUsersByRole_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_e")
	repo := NewRepository(db, nil)

	// Create users with different roles
	roles := []string{"CAREGIVER", "CAREGIVER", "MUNICIPALITY", "INSURER"}
	for i, role := range roles {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          "role" + string(rune('0'+i)) + "@test.com",
			FirstName:      "Role",
			LastName:       string(rune('A' + i)),
			Role:           role,
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
	}

	// List only caregivers
	users, total, err := repo.ListActiveUsersByRoleWithPagination(schemaName, "CAREGIVER", 10, 0, "")
	if err != nil {
		t.Fatalf("ListActiveUsersByRoleWithPagination failed: %v", err)
	}

	if total < 2 {
		t.Errorf("Expected at least 2 caregivers, got %d", total)
	}

	// Verify all returned users are caregivers
	for _, user := range users {
		if user.Role != "CAREGIVER" {
			t.Errorf("Expected role CAREGIVER, got %s", user.Role)
		}
	}
}

// TestRepositoryUpdate_Integration tests updating a user
func TestRepositoryUpdate_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_f")
	repo := NewRepository(db, nil)

	// Create user
	user := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "update@test.com",
		FirstName:      "Update",
		LastName:       "User",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update user
	user.Email = "updated@test.com"
	user.FirstName = "Updated"

	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(schemaName, user.ID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}

	if retrieved.Email != "updated@test.com" {
		t.Errorf("Expected email updated@test.com, got %s", retrieved.Email)
	}

	if retrieved.FirstName != "Updated" {
		t.Errorf("Expected first name Updated, got %s", retrieved.FirstName)
	}
}

// TestRepositoryDelete_Integration tests soft deleting a user
func TestRepositoryDelete_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_g")
	repo := NewRepository(db, nil)

	// Create user
	user := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "delete@test.com",
		FirstName:      "Delete",
		LastName:       "User",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete user
	err = repo.Delete(schemaName, orgID, user.ID, "CAREGIVER")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify user is soft deleted
	var deletedAt *string
	query := "SELECT deleted_at FROM " + schemaName + ".users WHERE id = $1"
	err = db.QueryRow(query, user.ID).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("Failed to query deleted user: %v", err)
	}

	if deletedAt == nil {
		t.Error("Expected deleted_at to be set after deletion")
	}
}

// TestRepositoryGetByKeycloakID_Integration tests getting user by Keycloak ID
func TestRepositoryGetByKeycloakID_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_h")
	repo := NewRepository(db, nil)

	keycloakID := uuid.New().String()

	// Create user
	user := &User{
		KeycloakUserID: keycloakID,
		Email:          "kc@test.com",
		FirstName:      "KC",
		LastName:       "User",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get user by Keycloak ID
	retrieved, err := repo.GetByKeycloakID(schemaName, keycloakID)
	if err != nil {
		t.Fatalf("GetByKeycloakID failed: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.KeycloakUserID != keycloakID {
		t.Errorf("Expected Keycloak ID %s, got %s", keycloakID, retrieved.KeycloakUserID)
	}
}

// TestRepositoryEmployeeIDGeneration_Integration tests sequential employee ID generation
func TestRepositoryEmployeeIDGeneration_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_i")
	repo := NewRepository(db, nil)

	// Create 3 users and verify sequential employee IDs
	expectedIDs := []string{"EMP-0001", "EMP-0002", "EMP-0003"}

	for i, expectedID := range expectedIDs {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          "emp" + string(rune('0'+i)) + "@test.com",
			FirstName:      "Employee",
			LastName:       string(rune('A' + i)),
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}

		if user.EmployeeID != expectedID {
			t.Errorf("Expected employee ID %s, got %s", expectedID, user.EmployeeID)
		}
	}
}

// TestRepositoryGetSchemaNameByOrgID_Integration tests getting schema name by org ID
func TestRepositoryGetSchemaNameByOrgID_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, expectedSchema := testutil.CreateTestOrg(t, db, "hospital_j")
	repo := NewRepository(db, nil)

	// Get schema name by org ID
	schemaName, err := repo.GetSchemaNameByOrgID(orgID)
	if err != nil {
		t.Fatalf("GetSchemaNameByOrgID failed: %v", err)
	}

	if schemaName != expectedSchema {
		t.Errorf("Expected schema %s, got %s", expectedSchema, schemaName)
	}
}

// TestRepositoryGetSchemaNameByOrgID_NotFound_Integration tests getting schema for non-existent org
func TestRepositoryGetSchemaNameByOrgID_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Try to get schema for non-existent org
	_, err := repo.GetSchemaNameByOrgID("00000000-0000-0000-0000-000000000000")
	if err != ErrInvalidOrgSchema {
		t.Errorf("Expected ErrInvalidOrgSchema, got %v", err)
	}
}

// TestRepositoryValidateOrgSchema_Integration tests schema validation
func TestRepositoryValidateOrgSchema_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	_, schemaName := testutil.CreateTestOrg(t, db, "hospital_k")
	repo := NewRepository(db, nil)

	// Valid schema
	err := repo.ValidateOrgSchema(schemaName)
	if err != nil {
		t.Errorf("Expected valid schema, got error: %v", err)
	}

	// Invalid schema
	err = repo.ValidateOrgSchema("nonexistent_schema")
	if err != ErrInvalidOrgSchema {
		t.Errorf("Expected ErrInvalidOrgSchema, got %v", err)
	}
}

// TestRepositoryListWithPagination_Search_Integration tests search functionality
func TestRepositoryListWithPagination_Search_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_l")
	repo := NewRepository(db, nil)

	// Create users with different names
	testUsers := []struct {
		firstName string
		lastName  string
		email     string
	}{
		{"Alice", "Smith", "alice.smith@test.com"},
		{"Bob", "Johnson", "bob.johnson@test.com"},
		{"Alice", "Brown", "alice.brown@test.com"},
		{"Charlie", "Smith", "charlie.smith@test.com"},
	}

	for _, tu := range testUsers {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          tu.email,
			FirstName:      tu.firstName,
			LastName:       tu.lastName,
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user failed: %v", err)
		}
	}

	// Search for "Alice"
	users, total, err := repo.ListWithPagination(schemaName, 10, 0, "Alice")
	if err != nil {
		t.Fatalf("Search for Alice failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 users matching 'Alice', got %d", total)
	}

	// Search for "Smith"
	users, total, err = repo.ListWithPagination(schemaName, 10, 0, "Smith")
	if err != nil {
		t.Fatalf("Search for Smith failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 users matching 'Smith', got %d", total)
	}

	// Search by email
	users, total, err = repo.ListWithPagination(schemaName, 10, 0, "bob.johnson")
	if err != nil {
		t.Fatalf("Search by email failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 user matching 'bob.johnson', got %d", total)
	}

	if len(users) > 0 && users[0].FirstName != "Bob" {
		t.Errorf("Expected to find Bob, got %s", users[0].FirstName)
	}
}

// TestRepositoryListActiveUsersByRole_Search_Integration tests role-based search
func TestRepositoryListActiveUsersByRole_Search_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_m")
	repo := NewRepository(db, nil)

	// Create caregivers and municipality users
	caregivers := []string{"Alice", "Bob", "Charlie"}
	for _, name := range caregivers {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          name + "@test.com",
			FirstName:      name,
			LastName:       "Caregiver",
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}
		repo.Create(user)
	}

	// Create municipality user
	user := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "alice@muni.com",
		FirstName:      "Alice",
		LastName:       "Municipality",
		Role:           "MUNICIPALITY",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}
	repo.Create(user)

	// Search for "Alice" among caregivers only
	users, total, err := repo.ListActiveUsersByRoleWithPagination(schemaName, "CAREGIVER", 10, 0, "Alice")
	if err != nil {
		t.Fatalf("Search caregivers failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 caregiver named Alice, got %d", total)
	}

	if len(users) > 0 && users[0].Role != "CAREGIVER" {
		t.Errorf("Expected CAREGIVER role, got %s", users[0].Role)
	}

	// Search for "Alice" among municipality users
	users, total, err = repo.ListActiveUsersByRoleWithPagination(schemaName, "MUNICIPALITY", 10, 0, "Alice")
	if err != nil {
		t.Fatalf("Search municipality failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 municipality user named Alice, got %d", total)
	}
}

// TestRepositorySoftDelete_ExcludesFromActiveList_Integration tests soft delete behavior
func TestRepositorySoftDelete_ExcludesFromActiveList_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_n")
	repo := NewRepository(db, nil)

	// Create 3 caregivers
	var userIDs []string
	for i := 1; i <= 3; i++ {
		user := &User{
			KeycloakUserID: uuid.New().String(),
			Email:          "del" + string(rune('0'+i)) + "@test.com",
			FirstName:      "Delete",
			LastName:       string(rune('A' + i - 1)),
			Role:           "CAREGIVER",
			OrgID:          orgID,
			OrgSchemaName:  schemaName,
		}

		err := repo.Create(user)
		if err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
		userIDs = append(userIDs, user.ID)
	}

	// Verify all 3 are in active list
	users, total, err := repo.ListActiveUsersByRoleWithPagination(schemaName, "CAREGIVER", 10, 0, "")
	if err != nil {
		t.Fatalf("List active users failed: %v", err)
	}

	if total < 3 {
		t.Errorf("Expected at least 3 active caregivers, got %d", total)
	}

	// Soft delete one user
	err = repo.Delete(schemaName, orgID, userIDs[0], "CAREGIVER")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify only 2 are in active list now
	users, total, err = repo.ListActiveUsersByRoleWithPagination(schemaName, "CAREGIVER", 10, 0, "")
	if err != nil {
		t.Fatalf("List active users after delete failed: %v", err)
	}

	// Should have 2 fewer than before
	if total < 2 {
		t.Errorf("Expected at least 2 active caregivers after deletion, got %d", total)
	}

	// Verify deleted user is not in the list
	for _, user := range users {
		if user.ID == userIDs[0] {
			t.Error("Deleted user should not appear in active users list")
		}
	}
}

// TestRepositoryGetByID_NotFound_Integration tests error handling for non-existent user
func TestRepositoryGetByID_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	_, schemaName := testutil.CreateTestOrg(t, db, "hospital_o")
	repo := NewRepository(db, nil)

	// Try to get non-existent user
	_, err := repo.GetByID(schemaName, "00000000-0000-0000-0000-000000000000")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

// TestRepositoryUpdate_NotFound_Integration tests updating non-existent user
func TestRepositoryUpdate_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	_, schemaName := testutil.CreateTestOrg(t, db, "hospital_p")
	repo := NewRepository(db, nil)

	user := &User{
		ID:            "00000000-0000-0000-0000-000000000000",
		Email:         "notfound@test.com",
		FirstName:     "Not",
		LastName:      "Found",
		OrgSchemaName: schemaName,
	}

	err := repo.Update(user)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

// TestRepositoryDelete_NotFound_Integration tests deleting non-existent user
func TestRepositoryDelete_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_q")
	repo := NewRepository(db, nil)

	err := repo.Delete(schemaName, orgID, "00000000-0000-0000-0000-000000000000", "CAREGIVER")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

// TestRepositoryDelete_AlreadyDeleted_Integration tests deleting already deleted user
func TestRepositoryDelete_AlreadyDeleted_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_r")
	repo := NewRepository(db, nil)

	// Create user
	user := &User{
		KeycloakUserID: uuid.New().String(),
		Email:          "double@test.com",
		FirstName:      "Double",
		LastName:       "Delete",
		Role:           "CAREGIVER",
		OrgID:          orgID,
		OrgSchemaName:  schemaName,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete once
	err = repo.Delete(schemaName, orgID, user.ID, "CAREGIVER")
	if err != nil {
		t.Fatalf("First delete failed: %v", err)
	}

	// Try to delete again
	err = repo.Delete(schemaName, orgID, user.ID, "CAREGIVER")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound on second delete, got %v", err)
	}
}
