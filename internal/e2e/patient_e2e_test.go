//go:build integration

package e2e

import (
	"net/http"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

// TestE2E_CreatePatient_FullFlow tests creating a patient
func TestE2E_CreatePatient_FullFlow(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization first
	orgBody := map[string]interface{}{
		"name":          "Hospital for Patient Test",
		"contact_email": "patienttest@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)
	orgID := orgResult.Organization.ID

	// Check events before patient creation
	eventsBefore := ts.MockPublisher.GetEventCountByKey("patient.created")

	// Create patient
	patientBody := map[string]interface{}{
		"username":    "patient1",
		"email":       "patient1@test.com",
		"firstName":   "John",
		"lastName":    "Doe",
		"dateOfBirth": "1980-01-15",
		"address":     "123 Patient St",
		"phoneNumber": "+1234567890",
		"temporaryPassword": "TempPass123!",
	}

	patientResp := client.POSTWithOrgHeader(t, "/organization/patients", patientBody, orgID)
	testutil.AssertStatusCode(t, patientResp, http.StatusCreated)

	var patientResult struct {
		Success bool `json:"success"`
		Patient struct {
			ID             string `json:"id"`
			PatientID      string `json:"patient_id"`
			KeycloakUserID string `json:"keycloak_user_id"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
			Email          string `json:"email"`
			IsActive       bool   `json:"is_active"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, patientResp, &patientResult)

	if !patientResult.Success {
		t.Error("Expected success to be true")
	}

	if patientResult.Patient.FirstName != "John" {
		t.Errorf("Expected firstName 'John', got '%s'", patientResult.Patient.FirstName)
	}

	if patientResult.Patient.PatientID == "" {
		t.Error("Expected patient ID to be generated")
	}

	if !patientResult.Patient.IsActive {
		t.Error("Expected patient to be active")
	}

	// Verify event was published
	ts.MockPublisher.AssertEventPublished(t, "patient.created")

	eventsAfter := ts.MockPublisher.GetEventCountByKey("patient.created")
	if eventsAfter != eventsBefore+1 {
		t.Errorf("Expected %d patient.created events, got %d", eventsBefore+1, eventsAfter)
	}

	t.Logf("E2E Test Passed: Created patient %s with event published", patientResult.Patient.ID)
}

// TestE2E_ListPatients_WithPagination tests listing patients with pagination
func TestE2E_ListPatients_WithPagination(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization
	orgBody := map[string]interface{}{
		"name":          "Hospital for List Patients",
		"contact_email": "listpatients@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)
	orgID := orgResult.Organization.ID

	// Create multiple patients
	for i := 0; i < 3; i++ {
		patientBody := map[string]interface{}{
			"username":          "patient" + string(rune('1'+i)),
			"email":             "patient" + string(rune('1'+i)) + "@test.com",
			"firstName":         "Patient",
			"lastName":          string(rune('A' + i)),
			"dateOfBirth":       "1980-01-01",
			"address":           "123 Test St",
			"temporaryPassword": "TempPass123!",
		}

		patientResp := client.POSTWithOrgHeader(t, "/organization/patients", patientBody, orgID)
		testutil.AssertStatusCode(t, patientResp, http.StatusCreated)
	}

	// Use ORG_ADMIN token to list patients (needs orgSchemaName)
	orgAdminToken := ts.GenerateOrgAdminToken(t, orgID, orgResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	// List patients
	listResp := orgAdminClient.GET(t, "/organization/patients?page=1&limit=10")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Success bool `json:"success"`
		Patients []struct {
			ID        string `json:"id"`
			FirstName string `json:"first_name"`
		} `json:"patients"`
		Pagination struct {
			CurrentPage  int `json:"current_page"`
			TotalRecords int `json:"total_records"`
		} `json:"pagination"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	if len(listResult.Patients) < 3 {
		t.Errorf("Expected at least 3 patients, got %d", len(listResult.Patients))
	}

	t.Logf("E2E Test Passed: Listed %d patients with pagination", len(listResult.Patients))
}

// TestE2E_ListActivePatients_FilterByStatus tests filtering active patients
func TestE2E_ListActivePatients_FilterByStatus(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization
	orgBody := map[string]interface{}{
		"name":          "Hospital for Active Patients",
		"contact_email": "activepatients@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)

	// Create ORG_ADMIN token
	orgAdminToken := ts.GenerateOrgAdminToken(t, orgResult.Organization.ID, orgResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	// List active patients
	listResp := orgAdminClient.GET(t, "/organization/patients/active?page=1&limit=10")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Success bool `json:"success"`
		Patients []struct {
			IsActive bool `json:"is_active"`
		} `json:"patients"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// All returned patients should be active
	for _, p := range listResult.Patients {
		if !p.IsActive {
			t.Error("Expected all patients to be active")
		}
	}

	t.Logf("E2E Test Passed: Listed %d active patients", len(listResult.Patients))
}

// TestE2E_UpdatePatient_FullFlow tests updating a patient
func TestE2E_UpdatePatient_FullFlow(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization
	orgBody := map[string]interface{}{
		"name":          "Hospital for Update Patient",
		"contact_email": "updatepatient@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)
	orgID := orgResult.Organization.ID

	// Create patient
	createBody := map[string]interface{}{
		"username":          "updatepatient",
		"email":             "original@patient.com",
		"firstName":         "Original",
		"lastName":          "Patient",
		"dateOfBirth":       "1985-03-10",
		"address":           "123 Original St",
		"temporaryPassword": "TempPass123!",
	}

	createResp := client.POSTWithOrgHeader(t, "/organization/patients", createBody, orgID)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Patient struct {
			ID string `json:"id"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	patientID := createResult.Patient.ID

	// Update patient using ORG_ADMIN token
	orgAdminToken := ts.GenerateOrgAdminToken(t, orgID, orgResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	updateBody := map[string]interface{}{
		"email":   "updated@patient.com",
		"address": "456 Updated Ave",
	}

	updateResp := orgAdminClient.PUT(t, "/organization/patients/"+patientID, updateBody)
	testutil.AssertStatusCode(t, updateResp, http.StatusOK)

	var updateResult struct {
		Success bool `json:"success"`
		Patient struct {
			Email   string `json:"email"`
			Address string `json:"address"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, updateResp, &updateResult)

	if updateResult.Patient.Email != "updated@patient.com" {
		t.Errorf("Expected email 'updated@patient.com', got '%s'", updateResult.Patient.Email)
	}

	if updateResult.Patient.Address != "456 Updated Ave" {
		t.Errorf("Expected address '456 Updated Ave', got '%s'", updateResult.Patient.Address)
	}

	t.Logf("E2E Test Passed: Updated patient %s successfully", patientID)
}

// TestE2E_DeletePatient_SoftDelete tests soft deleting a patient
func TestE2E_DeletePatient_SoftDelete(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization
	orgBody := map[string]interface{}{
		"name":          "Hospital for Delete Patient",
		"contact_email": "deletepatient@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)
	orgID := orgResult.Organization.ID

	// Check events before patient creation
	eventsBefore := ts.MockPublisher.GetEventCountByKey("patient.deleted")

	// Create patient
	createBody := map[string]interface{}{
		"username":          "deletepatient",
		"email":             "deletepatient@test.com",
		"firstName":         "Delete",
		"lastName":          "Patient",
		"dateOfBirth":       "1990-01-01",
		"address":           "789 Delete Rd",
		"temporaryPassword": "TempPass123!",
	}

	createResp := client.POSTWithOrgHeader(t, "/organization/patients", createBody, orgID)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Patient struct {
			ID             string `json:"id"`
			KeycloakUserID string `json:"keycloak_user_id"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	patientID := createResult.Patient.ID
	keycloakUserID := createResult.Patient.KeycloakUserID

	// Verify patient exists in Keycloak
	if !ts.MockKeycloak.UserExists(keycloakUserID) {
		t.Error("Patient should exist in mock Keycloak before deletion")
	}

	// Delete patient using ORG_ADMIN token
	orgAdminToken := ts.GenerateOrgAdminToken(t, orgID, orgResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	deleteResp := orgAdminClient.DELETE(t, "/organization/patients/"+patientID)
	testutil.AssertStatusCode(t, deleteResp, http.StatusOK)

	// Verify event was published
	ts.MockPublisher.AssertEventPublished(t, "patient.deleted")

	eventsAfter := ts.MockPublisher.GetEventCountByKey("patient.deleted")
	if eventsAfter != eventsBefore+1 {
		t.Errorf("Expected %d patient.deleted events, got %d", eventsBefore+1, eventsAfter)
	}

	// Verify soft delete in database
	var deletedAt *string
	err := ts.DB.QueryRow(`
		SELECT deleted_at 
		FROM "`+orgResult.Organization.SchemaName+`".patients 
		WHERE id = $1
	`, patientID).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("Failed to query patient from database: %v", err)
	}

	if deletedAt == nil {
		t.Error("Expected deleted_at to be set (soft delete)")
	}

	t.Logf("E2E Test Passed: Soft deleted patient %s with event published", patientID)
}

// TestE2E_GetPatient_ById tests retrieving a patient by ID
func TestE2E_GetPatient_ById(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create organization
	orgBody := map[string]interface{}{
		"name":          "Hospital for Get Patient",
		"contact_email": "getpatient@hospital.com",
	}

	orgResp := client.POST(t, "/organizations", orgBody)
	testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

	var orgResult struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, orgResp, &orgResult)
	orgID := orgResult.Organization.ID

	// Create patient
	createBody := map[string]interface{}{
		"username":          "getpatient",
		"email":             "getpatient@test.com",
		"firstName":         "Get",
		"lastName":          "Patient",
		"dateOfBirth":       "1990-05-20",
		"address":           "456 Get St",
		"temporaryPassword": "TempPass123!",
	}

	createResp := client.POSTWithOrgHeader(t, "/organization/patients", createBody, orgID)
	testutil.AssertStatusCode(t, createResp, http.StatusCreated)

	var createResult struct {
		Patient struct {
			ID string `json:"id"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, createResp, &createResult)
	patientID := createResult.Patient.ID

	// Get patient by ID
	orgAdminToken := ts.GenerateOrgAdminToken(t, orgID, orgResult.Organization.SchemaName)
	orgAdminClient := ts.NewClient(orgAdminToken)

	getResp := orgAdminClient.GET(t, "/organization/patients/"+patientID)
	testutil.AssertStatusCode(t, getResp, http.StatusOK)

	var getResult struct {
		Success bool `json:"success"`
		Patient struct {
			ID        string `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, getResp, &getResult)

	if getResult.Patient.ID != patientID {
		t.Errorf("Expected patient ID '%s', got '%s'", patientID, getResult.Patient.ID)
	}

	if getResult.Patient.FirstName != "Get" {
		t.Errorf("Expected firstName 'Get', got '%s'", getResult.Patient.FirstName)
	}

	t.Logf("E2E Test Passed: Retrieved patient %s successfully", patientID)
}
