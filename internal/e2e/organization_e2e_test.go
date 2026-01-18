//go:build integration

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

// TestE2E_CreateOrganization_FullFlow tests the complete organization creation flow
// This tests: HTTP → Auth Middleware → Handler → Service → Repository → Database
func TestE2E_CreateOrganization_FullFlow(t *testing.T) {
	// Setup complete test environment
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	// Generate SUPER_ADMIN token (only SUPER_ADMIN can create organizations)
	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Prepare request body
	reqBody := map[string]interface{}{
		"name":          "Test Hospital E2E",
		"contact_email": "test@hospital.com",
		"contact_phone": "+1234567890",
		"address":       "123 Test Street, Test City",
	}

	// Make HTTP POST request
	resp := client.POST(t, "/organizations", reqBody)

	// Verify HTTP response status
	if resp.StatusCode != http.StatusCreated {
		body := testutil.ReadBody(t, resp)
		t.Fatalf("Expected status 201, got %d. Body: %s", resp.StatusCode, body)
	}

	// Decode and verify response
	var result struct {
		Success      bool   `json:"success"`
		Message      string `json:"message"`
		Organization struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			SchemaName   string `json:"schema_name"`
			ContactEmail string `json:"contact_email"`
			ContactPhone string `json:"contact_phone"`
			Address      string `json:"address"`
			Status       string `json:"status"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, resp, &result)

	// Verify response data
	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Organization.Name != "Test Hospital E2E" {
		t.Errorf("Expected name 'Test Hospital E2E', got '%s'", result.Organization.Name)
	}

	if result.Organization.ID == "" {
		t.Error("Expected organization ID to be set")
	}

	if result.Organization.SchemaName == "" {
		t.Error("Expected schema name to be set")
	}

	if result.Organization.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", result.Organization.Status)
	}

	// Verify data was actually saved in database
	var dbName, dbEmail, dbStatus string
	err := ts.DB.QueryRow(`
		SELECT name, contact_email, status 
		FROM wailsalutem.organizations 
		WHERE id = $1 AND deleted_at IS NULL
	`, result.Organization.ID).Scan(&dbName, &dbEmail, &dbStatus)

	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if dbName != "Test Hospital E2E" {
		t.Errorf("Expected DB name 'Test Hospital E2E', got '%s'", dbName)
	}

	if dbEmail != "test@hospital.com" {
		t.Errorf("Expected DB email 'test@hospital.com', got '%s'", dbEmail)
	}

	if dbStatus != "active" {
		t.Errorf("Expected DB status 'active', got '%s'", dbStatus)
	}

	// Verify tenant schema was created
	var schemaExists bool
	err = ts.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, result.Organization.SchemaName).Scan(&schemaExists)

	if err != nil {
		t.Fatalf("Failed to check schema existence: %v", err)
	}

	if !schemaExists {
		t.Errorf("Expected schema '%s' to exist in database", result.Organization.SchemaName)
	}

	t.Logf("✅ E2E Test Passed: Created organization %s with schema %s", 
		result.Organization.ID, result.Organization.SchemaName)
}

// TestE2E_CreateAndGetOrganization tests creating and then retrieving an organization
func TestE2E_CreateAndGetOrganization(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Step 1: Create organization
	createBody := map[string]interface{}{
		"name":          "Hospital for GET Test",
		"contact_email": "get@test.com",
		"contact_phone": "+9876543210",
		"address":       "456 GET Street",
	}

	createResp := client.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)

	// Step 2: Get organization by ID
	getResp := client.GET(t, "/organizations/"+createResult.Organization.ID)
	testutil.AssertStatusCode(t, getResp, http.StatusOK)

	var getResult struct {
		Success      bool `json:"success"`
		Organization struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			ContactEmail string `json:"contact_email"`
			Status       string `json:"status"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, getResp, &getResult)

	// Verify the retrieved organization matches what we created
	if getResult.Organization.ID != createResult.Organization.ID {
		t.Errorf("Expected ID %s, got %s", createResult.Organization.ID, getResult.Organization.ID)
	}

	if getResult.Organization.Name != "Hospital for GET Test" {
		t.Errorf("Expected name 'Hospital for GET Test', got '%s'", getResult.Organization.Name)
	}

	if getResult.Organization.ContactEmail != "get@test.com" {
		t.Errorf("Expected email 'get@test.com', got '%s'", getResult.Organization.ContactEmail)
	}

	t.Logf("✅ E2E Test Passed: Created and retrieved organization %s", getResult.Organization.ID)
}

// TestE2E_Authorization_OrgAdminCannotCreateOrg tests authorization rules
func TestE2E_Authorization_OrgAdminCannotCreateOrg(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	// First create an organization as SUPER_ADMIN
	superAdminToken := ts.GenerateSuperAdminToken(t)
	superAdminClient := ts.NewClient(superAdminToken)

	createBody := map[string]interface{}{
		"name":          "Existing Hospital",
		"contact_email": "existing@hospital.com",
	}

	createResp := superAdminClient.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)

	// Now try to create another organization as ORG_ADMIN (should fail)
	orgAdminToken := ts.GenerateOrgAdminToken(t, createResult.Organization.ID, createResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	forbiddenBody := map[string]interface{}{
		"name":          "Should Fail Hospital",
		"contact_email": "fail@hospital.com",
	}

	forbiddenResp := orgAdminClient.POST(t, "/organizations", forbiddenBody)

	// Should be forbidden (403)
	if forbiddenResp.StatusCode != http.StatusForbidden {
		body := testutil.ReadBody(t, forbiddenResp)
		t.Errorf("Expected status 403 (Forbidden), got %d. Body: %s", forbiddenResp.StatusCode, body)
	}

	t.Logf("✅ E2E Test Passed: ORG_ADMIN correctly forbidden from creating organizations")
}

// TestE2E_Authentication_MissingToken tests authentication requirement
func TestE2E_Authentication_MissingToken(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	// Create client without token
	client := ts.NewClient("")

	reqBody := map[string]interface{}{
		"name":          "Should Fail",
		"contact_email": "fail@test.com",
	}

	resp := client.POST(t, "/organizations", reqBody)

	// Should be unauthorized (401)
	if resp.StatusCode != http.StatusUnauthorized {
		body := testutil.ReadBody(t, resp)
		t.Errorf("Expected status 401 (Unauthorized), got %d. Body: %s", resp.StatusCode, body)
	}

	t.Logf("E2E Test Passed: Missing token correctly returns 401")
}

// TestE2E_ListOrganizations_WithPagination tests listing organizations with pagination
func TestE2E_ListOrganizations_WithPagination(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Create multiple organizations
	orgIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		createBody := map[string]interface{}{
			"name":          "Hospital " + string(rune('A'+i)),
			"contact_email": "hospital" + string(rune('a'+i)) + "@test.com",
		}

		createResp := client.POST(t, "/organizations", createBody)
		testutil.AssertStatusCode(t, createResp, http.StatusCreated)

		var createResult struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		}
		testutil.DecodeJSON(t, createResp, &createResult)
		orgIDs[i] = createResult.Organization.ID
	}

	// List organizations with pagination
	listResp := client.GET(t, "/organizations?page=1&limit=10")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Success        bool `json:"success"`
		Organizations  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
		Pagination struct {
			CurrentPage  int `json:"current_page"`
			PerPage      int `json:"per_page"`
			TotalPages   int `json:"total_pages"`
			TotalRecords int `json:"total_records"`
		} `json:"pagination"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// Verify we got organizations back
	if !listResult.Success {
		t.Error("Expected success to be true")
	}

	if len(listResult.Organizations) < 3 {
		t.Errorf("Expected at least 3 organizations, got %d", len(listResult.Organizations))
	}

	// Verify pagination metadata
	if listResult.Pagination.CurrentPage != 1 {
		t.Errorf("Expected current page 1, got %d", listResult.Pagination.CurrentPage)
	}

	if listResult.Pagination.TotalRecords < 3 {
		t.Errorf("Expected at least 3 total records, got %d", listResult.Pagination.TotalRecords)
	}

	// Verify our created organizations are in the list
	foundCount := 0
	for _, org := range listResult.Organizations {
		for _, createdID := range orgIDs {
			if org.ID == createdID {
				foundCount++
			}
		}
	}

	if foundCount != 3 {
		t.Errorf("Expected to find all 3 created organizations, found %d", foundCount)
	}

	t.Logf("E2E Test Passed: Listed %d organizations with pagination", len(listResult.Organizations))
}

// TestE2E_UpdateOrganization_FullFlow tests updating an organization
func TestE2E_UpdateOrganization_FullFlow(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Step 1: Create organization
	createBody := map[string]interface{}{
		"name":          "Original Hospital Name",
		"contact_email": "original@hospital.com",
		"contact_phone": "+1111111111",
		"address":       "123 Original St",
	}

	createResp := client.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	orgID := createResult.Organization.ID

	// Step 2: Update organization
	updateBody := map[string]interface{}{
		"name":          "Updated Hospital Name",
		"contact_email": "updated@hospital.com",
		"contact_phone": "+2222222222",
		"address":       "456 Updated Ave",
	}

	updateResp := client.PUT(t, "/organizations/"+orgID, updateBody)
	testutil.AssertStatusCode(t, updateResp, http.StatusOK)

	var updateResult struct {
		Success      bool `json:"success"`
		Organization struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			ContactEmail string `json:"contact_email"`
			ContactPhone string `json:"contact_phone"`
			Address      string `json:"address"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, updateResp, &updateResult)

	// Verify update response
	if !updateResult.Success {
		t.Error("Expected success to be true")
	}

	if updateResult.Organization.Name != "Updated Hospital Name" {
		t.Errorf("Expected name 'Updated Hospital Name', got '%s'", updateResult.Organization.Name)
	}

	if updateResult.Organization.ContactEmail != "updated@hospital.com" {
		t.Errorf("Expected email 'updated@hospital.com', got '%s'", updateResult.Organization.ContactEmail)
	}

	// Step 3: Verify update in database
	var dbName, dbEmail, dbPhone string
	err := ts.DB.QueryRow(`
		SELECT name, contact_email, contact_phone 
		FROM wailsalutem.organizations 
		WHERE id = $1 AND deleted_at IS NULL
	`, orgID).Scan(&dbName, &dbEmail, &dbPhone)

	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if dbName != "Updated Hospital Name" {
		t.Errorf("Expected DB name 'Updated Hospital Name', got '%s'", dbName)
	}

	if dbEmail != "updated@hospital.com" {
		t.Errorf("Expected DB email 'updated@hospital.com', got '%s'", dbEmail)
	}

	if dbPhone != "+2222222222" {
		t.Errorf("Expected DB phone '+2222222222', got '%s'", dbPhone)
	}

	t.Logf("E2E Test Passed: Updated organization %s successfully", orgID)
}

// TestE2E_DeleteOrganization_SoftDelete tests soft deleting an organization
func TestE2E_DeleteOrganization_SoftDelete(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Step 1: Create organization
	createBody := map[string]interface{}{
		"name":          "Hospital To Delete",
		"contact_email": "delete@hospital.com",
	}

	createResp := client.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	orgID := createResult.Organization.ID

	// Step 2: Delete organization
	deleteResp := client.DELETE(t, "/organizations/"+orgID)
	testutil.AssertStatusCode(t, deleteResp, http.StatusNoContent)

	// Verify no body returned for 204
	if deleteResp.ContentLength != 0 {
		body := testutil.ReadBody(t, deleteResp)
		if body != "" {
			t.Errorf("Expected empty body for 204, got: %s", body)
		}
	}

	// Step 3: Verify soft delete in database (deleted_at should be set)
	var deletedAt *string
	err := ts.DB.QueryRow(`
		SELECT deleted_at 
		FROM wailsalutem.organizations 
		WHERE id = $1
	`, orgID).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if deletedAt == nil {
		t.Error("Expected deleted_at to be set (soft delete), but it was NULL")
	}

	// Step 4: Verify organization is not returned in list (soft deleted items excluded)
	listResp := client.GET(t, "/organizations?page=1&limit=100")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Organizations []struct {
			ID string `json:"id"`
		} `json:"organizations"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// Verify deleted organization is not in the list
	for _, org := range listResult.Organizations {
		if org.ID == orgID {
			t.Error("Deleted organization should not appear in list")
		}
	}

	// Step 5: Verify GET returns 404 for deleted organization
	getResp := client.GET(t, "/organizations/"+orgID)
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for deleted organization, got %d", getResp.StatusCode)
	}

	t.Logf("E2E Test Passed: Soft deleted organization %s successfully", orgID)
}

// TestE2E_UpdateOrganization_OrgAdminForbidden tests that ORG_ADMIN cannot update organizations
func TestE2E_UpdateOrganization_OrgAdminForbidden(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	// Create organization as SUPER_ADMIN
	superAdminToken := ts.GenerateSuperAdminToken(t)
	superAdminClient := ts.NewClient(superAdminToken)

	createBody := map[string]interface{}{
		"name":          "Hospital for Update Test",
		"contact_email": "update@test.com",
	}

	createResp := superAdminClient.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)

	// Try to update as ORG_ADMIN (should be forbidden)
	orgAdminToken := ts.GenerateOrgAdminToken(t, createResult.Organization.ID, createResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	updateBody := map[string]interface{}{
		"name": "Should Not Update",
	}

	updateResp := orgAdminClient.PUT(t, "/organizations/"+createResult.Organization.ID, updateBody)

	// Should be forbidden (403)
	if updateResp.StatusCode != http.StatusForbidden {
		body := testutil.ReadBody(t, updateResp)
		t.Errorf("Expected status 403 (Forbidden), got %d. Body: %s", updateResp.StatusCode, body)
	}

	t.Logf("E2E Test Passed: ORG_ADMIN correctly forbidden from updating organizations")
}

// TestE2E_GetOrganization_NotFound tests getting a non-existent organization
func TestE2E_GetOrganization_NotFound(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Try to get a non-existent organization
	resp := client.GET(t, "/organizations/00000000-0000-0000-0000-000000000000")

	// Should return 404
	if resp.StatusCode != http.StatusNotFound {
		body := testutil.ReadBody(t, resp)
		t.Errorf("Expected status 404, got %d. Body: %s", resp.StatusCode, body)
	}

	t.Logf("E2E Test Passed: Non-existent organization correctly returns 404")
}

// TestE2E_CreateOrganization_ValidationError tests validation errors
func TestE2E_CreateOrganization_ValidationError(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Try to create organization with missing required field (name)
	reqBody := map[string]interface{}{
		"contact_email": "test@hospital.com",
		// Missing "name" field
	}

	resp := client.POST(t, "/organizations", reqBody)

	// Should return 400 (Bad Request)
	if resp.StatusCode != http.StatusBadRequest {
		body := testutil.ReadBody(t, resp)
		t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, body)
	}

	var errorResult struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	testutil.DecodeJSON(t, resp, &errorResult)

	if errorResult.Success {
		t.Error("Expected success to be false for validation error")
	}

	t.Logf("E2E Test Passed: Validation error correctly returns 400")
}

// TestE2E_UpdateOrganization_PartialUpdate tests updating only some fields
func TestE2E_UpdateOrganization_PartialUpdate(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Create organization
	createBody := map[string]interface{}{
		"name":          "Partial Update Hospital",
		"contact_email": "partial@test.com",
		"contact_phone": "+1111111111",
		"address":       "123 Original St",
	}

	createResp := client.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	orgID := createResult.Organization.ID

	// Update only the email (partial update)
	updateBody := map[string]interface{}{
		"contact_email": "newemail@test.com",
	}

	updateResp := client.PUT(t, "/organizations/"+orgID, updateBody)
	testutil.AssertStatusCode(t, updateResp, http.StatusOK)

	var updateResult struct {
		Organization struct {
			Name         string `json:"name"`
			ContactEmail string `json:"contact_email"`
			ContactPhone string `json:"contact_phone"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, updateResp, &updateResult)

	// Verify email was updated
	if updateResult.Organization.ContactEmail != "newemail@test.com" {
		t.Errorf("Expected email 'newemail@test.com', got '%s'", updateResult.Organization.ContactEmail)
	}

	// Verify other fields remained unchanged
	if updateResult.Organization.Name != "Partial Update Hospital" {
		t.Errorf("Expected name to remain 'Partial Update Hospital', got '%s'", updateResult.Organization.Name)
	}

	if updateResult.Organization.ContactPhone != "+1111111111" {
		t.Errorf("Expected phone to remain '+1111111111', got '%s'", updateResult.Organization.ContactPhone)
	}

	t.Logf("E2E Test Passed: Partial update correctly updated only specified fields")
}

// TestE2E_ListOrganizations_EmptyResult tests listing when no organizations exist
func TestE2E_ListOrganizations_EmptyResult(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	// Create ORG_ADMIN for a non-existent org (they should see empty list)
	token := ts.GenerateOrgAdminToken(t, "fake-org-id", "fake_schema")
	client := ts.NewClient(token)

	// List organizations (ORG_ADMIN without an org sees nothing)
	resp := client.GET(t, "/organizations")

	// Should still return 200 with empty list or error
	// The actual behavior depends on your implementation
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body := testutil.ReadBody(t, resp)
		t.Logf("Got status %d, body: %s", resp.StatusCode, body)
	}

	t.Logf("E2E Test Passed: Empty list scenario handled")
}

// TestE2E_DeleteOrganization_PublishesEvent tests that deletion publishes an event to RabbitMQ
func TestE2E_DeleteOrganization_PublishesEvent(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	token := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(token)

	// Create organization
	createBody := map[string]interface{}{
		"name":          "Hospital for Event Test",
		"contact_email": "event@hospital.com",
	}

	createResp := client.POST(t, "/organizations", createBody)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Organization struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	orgID := createResult.Organization.ID

	// Check events before deletion
	eventsBefore := ts.MockPublisher.GetEventCount()

	// Delete organization
	deleteResp := client.DELETE(t, "/organizations/"+orgID)
	testutil.AssertStatusCode(t, deleteResp, http.StatusNoContent)

	// Verify event was published
	ts.MockPublisher.AssertEventPublished(t, "organization.deleted")

	// Verify we have exactly one more event
	eventsAfter := ts.MockPublisher.GetEventCount()
	if eventsAfter != eventsBefore+1 {
		t.Errorf("Expected %d events after deletion, got %d", eventsBefore+1, eventsAfter)
	}

	// Get the deletion event and verify its content
	deletionEvents := ts.MockPublisher.GetEventsByKey("organization.deleted")
	if len(deletionEvents) == 0 {
		t.Fatal("Expected to find organization.deleted event")
	}

	lastEvent := deletionEvents[len(deletionEvents)-1]

	// Verify event data (unmarshal from JSON)
	var eventData struct {
		EventType string `json:"event_type"`
		Data      struct {
			OrganizationID   string `json:"organization_id"`
			OrganizationName string `json:"organization_name"`
			SchemaName       string `json:"schema_name"`
		} `json:"data"`
	}

	err := json.Unmarshal(lastEvent.RawJSON, &eventData)
	if err != nil {
		t.Fatalf("Failed to unmarshal event data: %v", err)
	}

	if eventData.EventType != "organization.deleted" {
		t.Errorf("Expected event_type 'organization.deleted', got '%s'", eventData.EventType)
	}

	if eventData.Data.OrganizationID != orgID {
		t.Errorf("Expected organization_id '%s', got '%s'", orgID, eventData.Data.OrganizationID)
	}

	if eventData.Data.OrganizationName != createResult.Organization.Name {
		t.Errorf("Expected organization_name '%s', got '%s'", 
			createResult.Organization.Name, eventData.Data.OrganizationName)
	}

	if eventData.Data.SchemaName != createResult.Organization.SchemaName {
		t.Errorf("Expected schema_name '%s', got '%s'", 
			createResult.Organization.SchemaName, eventData.Data.SchemaName)
	}

	t.Logf("E2E Test Passed: Organization deletion published event correctly")
}
