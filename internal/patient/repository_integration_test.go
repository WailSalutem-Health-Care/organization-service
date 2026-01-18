// go:build integration
//go:build integration

package patient

import (
	"context"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
	"github.com/google/uuid"
)

// TestRepositoryCreatePatient_Integration tests creating a patient in tenant schema
func TestRepositoryCreatePatient_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_a")
	repo := NewRepository(db, nil)

	req := CreatePatientRequest{
		FirstName:             "John",
		LastName:              "Doe",
		Email:                 "john.doe@example.com",
		PhoneNumber:           "+1234567890",
		DateOfBirth:           "1990-01-15",
		Address:               "123 Main St, City, State",
		EmergencyContactName:  "Jane Doe",
		EmergencyContactPhone: "+0987654321",
		MedicalNotes:          "No known allergies",
		CareplanType:          "basic",
		CareplanFrequency:     "weekly",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	if patient.ID == "" {
		t.Error("Expected patient ID to be set")
	}

	if patient.PatientID != "PT-0001" {
		t.Errorf("Expected first patient ID to be PT-0001, got %s", patient.PatientID)
	}

	if patient.FirstName != req.FirstName {
		t.Errorf("Expected first name %s, got %s", req.FirstName, patient.FirstName)
	}

	if patient.IsActive != true {
		t.Error("Expected patient to be active by default")
	}

	if patient.CareplanType != "basic" {
		t.Errorf("Expected careplan type basic, got %s", patient.CareplanType)
	}
}

// TestRepositoryGetPatient_Integration tests retrieving a patient
func TestRepositoryGetPatient_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_b")
	repo := NewRepository(db, nil)

	// Create patient
	req := CreatePatientRequest{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		DateOfBirth: "1985-05-20",
		Address:     "456 Oak Ave",
	}

	created, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	// Get patient
	patient, err := repo.GetPatient(context.Background(), schemaName, created.ID)
	if err != nil {
		t.Fatalf("GetPatient failed: %v", err)
	}

	if patient.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, patient.ID)
	}

	if patient.FirstName != "Alice" {
		t.Errorf("Expected first name Alice, got %s", patient.FirstName)
	}

	if patient.PatientID != "PT-0001" {
		t.Errorf("Expected patient ID PT-0001, got %s", patient.PatientID)
	}
}

// TestRepositoryGetPatient_CrossTenantIsolation_Integration tests tenant isolation
func TestRepositoryGetPatient_CrossTenantIsolation_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	// Create two organizations
	org1ID, schema1 := testutil.CreateTestOrg(t, db, "hospital_1")
	org2ID, schema2 := testutil.CreateTestOrg(t, db, "hospital_2")

	repo := NewRepository(db, nil)

	// Create patient in org1
	req1 := CreatePatientRequest{
		FirstName:   "Patient",
		LastName:    "One",
		Email:       "patient1@org1.com",
		DateOfBirth: "1990-01-01",
		Address:     "Org1 Address",
	}

	patient1, err := repo.CreatePatient(context.Background(), schema1, org1ID, uuid.New().String(), req1)
	if err != nil {
		t.Fatalf("Create patient in org1 failed: %v", err)
	}

	// Try to get patient1 from org2 schema (should fail)
	_, err = repo.GetPatient(context.Background(), schema2, patient1.ID)
	if err == nil {
		t.Error("Expected error when accessing patient from different tenant, got nil")
	}

	// Create patient in org2
	req2 := CreatePatientRequest{
		FirstName:   "Patient",
		LastName:    "Two",
		Email:       "patient2@org2.com",
		DateOfBirth: "1992-02-02",
		Address:     "Org2 Address",
	}

	patient2, err := repo.CreatePatient(context.Background(), schema2, org2ID, uuid.New().String(), req2)
	if err != nil {
		t.Fatalf("Create patient in org2 failed: %v", err)
	}

	// Verify patient2 is accessible from org2
	retrieved, err := repo.GetPatient(context.Background(), schema2, patient2.ID)
	if err != nil {
		t.Fatalf("GetPatient from own tenant failed: %v", err)
	}

	if retrieved.ID != patient2.ID {
		t.Errorf("Expected patient ID %s, got %s", patient2.ID, retrieved.ID)
	}
}

// TestRepositoryListPatients_Integration tests listing patients
func TestRepositoryListPatients_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_c")
	repo := NewRepository(db, nil)

	// Create multiple patients
	for i := 1; i <= 3; i++ {
		req := CreatePatientRequest{
			FirstName:   "Patient",
			LastName:    string(rune('A' + i - 1)),
			Email:       "patient" + string(rune('0'+i)) + "@test.com",
			DateOfBirth: "1990-01-01",
			Address:     "Address " + string(rune('0'+i)),
		}

		_, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient %d failed: %v", i, err)
		}
	}

	// List patients
	patients, err := repo.ListPatients(context.Background(), schemaName)
	if err != nil {
		t.Fatalf("ListPatients failed: %v", err)
	}

	if len(patients) < 3 {
		t.Errorf("Expected at least 3 patients, got %d", len(patients))
	}
}

// TestRepositoryListPatientsWithPagination_Integration tests paginated patient listing
func TestRepositoryListPatientsWithPagination_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_d")
	repo := NewRepository(db, nil)

	// Create 5 patients
	for i := 1; i <= 5; i++ {
		req := CreatePatientRequest{
			FirstName:   "Page",
			LastName:    string(rune('A' + i - 1)),
			Email:       "page" + string(rune('0'+i)) + "@test.com",
			DateOfBirth: "1990-01-01",
			Address:     "Address " + string(rune('0'+i)),
		}

		_, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient %d failed: %v", i, err)
		}
	}

	// Get first page (limit 2)
	patients, total, err := repo.ListPatientsWithPagination(context.Background(), schemaName, 2, 0, "")
	if err != nil {
		t.Fatalf("ListPatientsWithPagination failed: %v", err)
	}

	if len(patients) != 2 {
		t.Errorf("Expected 2 patients on first page, got %d", len(patients))
	}

	if total < 5 {
		t.Errorf("Expected total >= 5, got %d", total)
	}

	// Get second page
	patients, _, err = repo.ListPatientsWithPagination(context.Background(), schemaName, 2, 2, "")
	if err != nil {
		t.Fatalf("ListPatientsWithPagination page 2 failed: %v", err)
	}

	if len(patients) != 2 {
		t.Errorf("Expected 2 patients on second page, got %d", len(patients))
	}
}

// TestRepositoryListPatientsWithPagination_Search_Integration tests search functionality
func TestRepositoryListPatientsWithPagination_Search_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_e")
	repo := NewRepository(db, nil)

	// Create patients with different names
	testPatients := []struct {
		firstName string
		lastName  string
		email     string
	}{
		{"Alice", "Johnson", "alice.johnson@test.com"},
		{"Bob", "Smith", "bob.smith@test.com"},
		{"Alice", "Brown", "alice.brown@test.com"},
		{"Charlie", "Johnson", "charlie.johnson@test.com"},
	}

	for _, tp := range testPatients {
		req := CreatePatientRequest{
			FirstName:   tp.firstName,
			LastName:    tp.lastName,
			Email:       tp.email,
			DateOfBirth: "1990-01-01",
			Address:     "Test Address",
		}

		_, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient failed: %v", err)
		}
	}

	// Search for "Alice"
	patients, total, err := repo.ListPatientsWithPagination(context.Background(), schemaName, 10, 0, "Alice")
	if err != nil {
		t.Fatalf("Search for Alice failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 patients matching 'Alice', got %d", total)
	}

	// Search for "Johnson"
	patients, total, err = repo.ListPatientsWithPagination(context.Background(), schemaName, 10, 0, "Johnson")
	if err != nil {
		t.Fatalf("Search for Johnson failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 patients matching 'Johnson', got %d", total)
	}

	// Search by email
	patients, total, err = repo.ListPatientsWithPagination(context.Background(), schemaName, 10, 0, "bob.smith")
	if err != nil {
		t.Fatalf("Search by email failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 patient matching 'bob.smith', got %d", total)
	}

	if len(patients) > 0 && patients[0].FirstName != "Bob" {
		t.Errorf("Expected to find Bob, got %s", patients[0].FirstName)
	}
}

// TestRepositoryListActivePatientsWithPagination_Integration tests active patient filtering
func TestRepositoryListActivePatientsWithPagination_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_f")
	repo := NewRepository(db, nil)

	// Create 3 patients
	for i := 1; i <= 3; i++ {
		req := CreatePatientRequest{
			FirstName:   "Active",
			LastName:    string(rune('A' + i - 1)),
			Email:       "active" + string(rune('0'+i)) + "@test.com",
			DateOfBirth: "1990-01-01",
			Address:     "Address " + string(rune('0'+i)),
		}

		_, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient %d failed: %v", i, err)
		}
	}

	// List active patients
	patients, total, err := repo.ListActivePatientsWithPagination(context.Background(), schemaName, 10, 0, "")
	if err != nil {
		t.Fatalf("ListActivePatientsWithPagination failed: %v", err)
	}

	if total < 3 {
		t.Errorf("Expected at least 3 active patients, got %d", total)
	}

	// Verify all returned patients are active
	for _, patient := range patients {
		if !patient.IsActive {
			t.Errorf("Expected all patients to be active, got inactive patient %s", patient.ID)
		}
	}
}

// TestRepositoryUpdate_Integration tests updating a patient
func TestRepositoryUpdate_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_g")
	repo := NewRepository(db, nil)

	// Create patient
	req := CreatePatientRequest{
		FirstName:   "Update",
		LastName:    "Test",
		Email:       "update@test.com",
		DateOfBirth: "1990-01-01",
		Address:     "Old Address",
		CareplanType: "basic",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	// Update patient
	newFirstName := "Updated"
	newAddress := "New Address"
	newCareplan := "intensive"
	updateReq := UpdatePatientRequest{
		FirstName:    &newFirstName,
		Address:      &newAddress,
		CareplanType: &newCareplan,
	}

	updated, err := repo.UpdatePatient(context.Background(), schemaName, patient.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdatePatient failed: %v", err)
	}

	if updated.FirstName != "Updated" {
		t.Errorf("Expected first name Updated, got %s", updated.FirstName)
	}

	if updated.Address != "New Address" {
		t.Errorf("Expected address 'New Address', got %s", updated.Address)
	}

	if updated.CareplanType != "intensive" {
		t.Errorf("Expected careplan type intensive, got %s", updated.CareplanType)
	}

	// Verify LastName was not changed
	if updated.LastName != "Test" {
		t.Errorf("Expected last name to remain Test, got %s", updated.LastName)
	}
}

// TestRepositoryDelete_Integration tests soft deleting a patient
func TestRepositoryDelete_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_h")
	repo := NewRepository(db, nil)

	// Create patient
	req := CreatePatientRequest{
		FirstName:   "Delete",
		LastName:    "Test",
		Email:       "delete@test.com",
		DateOfBirth: "1990-01-01",
		Address:     "Test Address",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	// Delete patient
	err = repo.DeletePatient(context.Background(), schemaName, orgID, patient.ID)
	if err != nil {
		t.Fatalf("DeletePatient failed: %v", err)
	}

	// Verify patient is soft deleted (deleted_at is set)
	var deletedAt *string
	query := "SELECT deleted_at FROM " + schemaName + ".patients WHERE id = $1"
	err = db.QueryRow(query, patient.ID).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("Failed to query deleted patient: %v", err)
	}

	if deletedAt == nil {
		t.Error("Expected deleted_at to be set after deletion")
	}

	// Verify patient cannot be retrieved after deletion
	_, err = repo.GetPatient(context.Background(), schemaName, patient.ID)
	if err == nil {
		t.Error("Expected error when getting deleted patient, got nil")
	}
}

// TestRepositoryPatientIDGeneration_Integration tests sequential patient ID generation
func TestRepositoryPatientIDGeneration_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_i")
	repo := NewRepository(db, nil)

	// Create 3 patients and verify sequential patient IDs
	expectedIDs := []string{"PT-0001", "PT-0002", "PT-0003"}

	for i, expectedID := range expectedIDs {
		req := CreatePatientRequest{
			FirstName:   "Patient",
			LastName:    string(rune('A' + i)),
			Email:       "pt" + string(rune('0'+i)) + "@test.com",
			DateOfBirth: "1990-01-01",
			Address:     "Address " + string(rune('0'+i)),
		}

		patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient %d failed: %v", i, err)
		}

		if patient.PatientID != expectedID {
			t.Errorf("Expected patient ID %s, got %s", expectedID, patient.PatientID)
		}
	}
}

// TestRepositorySoftDelete_ExcludesFromList_Integration tests soft delete behavior
func TestRepositorySoftDelete_ExcludesFromList_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_j")
	repo := NewRepository(db, nil)

	// Create 3 patients
	var patientIDs []string
	for i := 1; i <= 3; i++ {
		req := CreatePatientRequest{
			FirstName:   "Delete",
			LastName:    string(rune('A' + i - 1)),
			Email:       "del" + string(rune('0'+i)) + "@test.com",
			DateOfBirth: "1990-01-01",
			Address:     "Address " + string(rune('0'+i)),
		}

		patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
		if err != nil {
			t.Fatalf("CreatePatient %d failed: %v", i, err)
		}
		patientIDs = append(patientIDs, patient.ID)
	}

	// Verify all 3 are in list
	patients, err := repo.ListPatients(context.Background(), schemaName)
	if err != nil {
		t.Fatalf("ListPatients failed: %v", err)
	}

	if len(patients) < 3 {
		t.Errorf("Expected at least 3 patients, got %d", len(patients))
	}

	// Soft delete one patient
	err = repo.DeletePatient(context.Background(), schemaName, orgID, patientIDs[0])
	if err != nil {
		t.Fatalf("DeletePatient failed: %v", err)
	}

	// Verify only 2 are in list now
	patients, err = repo.ListPatients(context.Background(), schemaName)
	if err != nil {
		t.Fatalf("ListPatients after delete failed: %v", err)
	}

	// Verify deleted patient is not in the list
	for _, patient := range patients {
		if patient.ID == patientIDs[0] {
			t.Error("Deleted patient should not appear in list")
		}
	}
}

// TestRepositoryGetPatient_NotFound_Integration tests error handling for non-existent patient
func TestRepositoryGetPatient_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	_, schemaName := testutil.CreateTestOrg(t, db, "hospital_k")
	repo := NewRepository(db, nil)

	// Try to get non-existent patient
	_, err := repo.GetPatient(context.Background(), schemaName, uuid.New().String())
	if err == nil {
		t.Error("Expected error for non-existent patient, got nil")
	}
}

// TestRepositoryUpdate_NotFound_Integration tests updating non-existent patient
func TestRepositoryUpdate_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	_, schemaName := testutil.CreateTestOrg(t, db, "hospital_l")
	repo := NewRepository(db, nil)

	newName := "Updated"
	updateReq := UpdatePatientRequest{
		FirstName: &newName,
	}

	_, err := repo.UpdatePatient(context.Background(), schemaName, uuid.New().String(), updateReq)
	if err == nil {
		t.Error("Expected error when updating non-existent patient, got nil")
	}
}

// TestRepositoryDelete_NotFound_Integration tests deleting non-existent patient
func TestRepositoryDelete_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_m")
	repo := NewRepository(db, nil)

	err := repo.DeletePatient(context.Background(), schemaName, orgID, uuid.New().String())
	if err == nil {
		t.Error("Expected error when deleting non-existent patient, got nil")
	}
}

// TestRepositoryDelete_AlreadyDeleted_Integration tests deleting already deleted patient
func TestRepositoryDelete_AlreadyDeleted_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_n")
	repo := NewRepository(db, nil)

	// Create patient
	req := CreatePatientRequest{
		FirstName:   "Double",
		LastName:    "Delete",
		Email:       "double@test.com",
		DateOfBirth: "1990-01-01",
		Address:     "Test Address",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	// Delete once
	err = repo.DeletePatient(context.Background(), schemaName, orgID, patient.ID)
	if err != nil {
		t.Fatalf("First delete failed: %v", err)
	}

	// Try to delete again
	err = repo.DeletePatient(context.Background(), schemaName, orgID, patient.ID)
	if err == nil {
		t.Error("Expected error on second delete, got nil")
	}
}

// TestRepositoryPatient_CareplanFields_Integration tests care plan data
func TestRepositoryPatient_CareplanFields_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_o")
	repo := NewRepository(db, nil)

	// Create patient with care plan
	req := CreatePatientRequest{
		FirstName:         "Care",
		LastName:          "Plan",
		Email:             "careplan@test.com",
		DateOfBirth:       "1990-01-01",
		Address:           "Test Address",
		CareplanType:      "intensive",
		CareplanFrequency: "daily",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	if patient.CareplanType != "intensive" {
		t.Errorf("Expected careplan type intensive, got %s", patient.CareplanType)
	}

	if patient.CareplanFrequency != "daily" {
		t.Errorf("Expected careplan frequency daily, got %s", patient.CareplanFrequency)
	}

	// Update care plan
	newType := "palliative"
	newFreq := "weekly"
	updateReq := UpdatePatientRequest{
		CareplanType:      &newType,
		CareplanFrequency: &newFreq,
	}

	updated, err := repo.UpdatePatient(context.Background(), schemaName, patient.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdatePatient failed: %v", err)
	}

	if updated.CareplanType != "palliative" {
		t.Errorf("Expected updated careplan type palliative, got %s", updated.CareplanType)
	}

	if updated.CareplanFrequency != "weekly" {
		t.Errorf("Expected updated careplan frequency weekly, got %s", updated.CareplanFrequency)
	}
}

// TestRepositoryPatient_EmergencyContact_Integration tests emergency contact data
func TestRepositoryPatient_EmergencyContact_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_p")
	repo := NewRepository(db, nil)

	// Create patient with emergency contact
	req := CreatePatientRequest{
		FirstName:             "Emergency",
		LastName:              "Contact",
		Email:                 "emergency@test.com",
		DateOfBirth:           "1990-01-01",
		Address:               "Test Address",
		EmergencyContactName:  "Jane Doe",
		EmergencyContactPhone: "+1234567890",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	if patient.EmergencyContactName != "Jane Doe" {
		t.Errorf("Expected emergency contact name Jane Doe, got %s", patient.EmergencyContactName)
	}

	if patient.EmergencyContactPhone != "+1234567890" {
		t.Errorf("Expected emergency contact phone +1234567890, got %s", patient.EmergencyContactPhone)
	}

	// Update emergency contact
	newName := "John Smith"
	newPhone := "+0987654321"
	updateReq := UpdatePatientRequest{
		EmergencyContactName:  &newName,
		EmergencyContactPhone: &newPhone,
	}

	updated, err := repo.UpdatePatient(context.Background(), schemaName, patient.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdatePatient failed: %v", err)
	}

	if updated.EmergencyContactName != "John Smith" {
		t.Errorf("Expected updated emergency contact name John Smith, got %s", updated.EmergencyContactName)
	}

	if updated.EmergencyContactPhone != "+0987654321" {
		t.Errorf("Expected updated emergency contact phone +0987654321, got %s", updated.EmergencyContactPhone)
	}
}

// TestRepositoryPatient_IsActiveFlag_Integration tests is_active flag behavior
func TestRepositoryPatient_IsActiveFlag_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	orgID, schemaName := testutil.CreateTestOrg(t, db, "hospital_q")
	repo := NewRepository(db, nil)

	// Create patient (active by default)
	req := CreatePatientRequest{
		FirstName:   "Active",
		LastName:    "Flag",
		Email:       "active@test.com",
		DateOfBirth: "1990-01-01",
		Address:     "Test Address",
	}

	patient, err := repo.CreatePatient(context.Background(), schemaName, orgID, uuid.New().String(), req)
	if err != nil {
		t.Fatalf("CreatePatient failed: %v", err)
	}

	if !patient.IsActive {
		t.Error("Expected patient to be active by default")
	}

	// Deactivate patient
	isActive := false
	updateReq := UpdatePatientRequest{
		IsActive: &isActive,
	}

	updated, err := repo.UpdatePatient(context.Background(), schemaName, patient.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdatePatient failed: %v", err)
	}

	if updated.IsActive {
		t.Error("Expected patient to be inactive after update")
	}

	// Verify inactive patient is excluded from active list
	activePatients, _, err := repo.ListActivePatientsWithPagination(context.Background(), schemaName, 10, 0, "")
	if err != nil {
		t.Fatalf("ListActivePatientsWithPagination failed: %v", err)
	}

	for _, p := range activePatients {
		if p.ID == patient.ID {
			t.Error("Inactive patient should not appear in active patients list")
		}
	}

	// Verify inactive patient is still in regular list (not deleted)
	allPatients, err := repo.ListPatients(context.Background(), schemaName)
	if err != nil {
		t.Fatalf("ListPatients failed: %v", err)
	}

	found := false
	for _, p := range allPatients {
		if p.ID == patient.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Inactive patient should still appear in regular list")
	}

	// Reactivate patient
	isActive = true
	updateReq = UpdatePatientRequest{
		IsActive: &isActive,
	}

	updated, err = repo.UpdatePatient(context.Background(), schemaName, patient.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdatePatient reactivation failed: %v", err)
	}

	if !updated.IsActive {
		t.Error("Expected patient to be active after reactivation")
	}

	// Verify reactivated patient appears in active list
	activePatients, _, err = repo.ListActivePatientsWithPagination(context.Background(), schemaName, 10, 0, "")
	if err != nil {
		t.Fatalf("ListActivePatientsWithPagination after reactivation failed: %v", err)
	}

	found = false
	for _, p := range activePatients {
		if p.ID == patient.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Reactivated patient should appear in active patients list")
	}
}
