//go:build integration

package e2e

import (
	"net/http"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

// TestE2E_MultiTenant_OrgAdminCanOnlySeeOwnOrg tests that ORG_ADMIN can only see their own organization
func TestE2E_MultiTenant_OrgAdminCanOnlySeeOwnOrg(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	superAdminClient := ts.NewClient(superAdminToken)

	// Create two organizations
	org1Body := map[string]interface{}{
		"name":          "Hospital A",
		"contact_email": "hospitala@test.com",
	}
	org1Resp := superAdminClient.POST(t, "/organizations", org1Body)
	testutil.AssertStatusCode(t, org1Resp, http.StatusCreated)

	var org1Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org1Resp, &org1Result)

	org2Body := map[string]interface{}{
		"name":          "Hospital B",
		"contact_email": "hospitalb@test.com",
	}
	org2Resp := superAdminClient.POST(t, "/organizations", org2Body)
	testutil.AssertStatusCode(t, org2Resp, http.StatusCreated)

	var org2Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org2Resp, &org2Result)

	// Create ORG_ADMIN token for Hospital A
	org1AdminToken := ts.GenerateOrgAdminToken(t, org1Result.Organization.ID, org1Result.Organization.SchemaName)
	org1AdminClient := ts.NewClient(org1AdminToken)

	// ORG_ADMIN from Hospital A lists organizations (should only see Hospital A)
	listResp := org1AdminClient.GET(t, "/organizations")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// Should only see one organization (their own)
	if len(listResult.Organizations) != 1 {
		t.Errorf("ORG_ADMIN should only see 1 organization, got %d", len(listResult.Organizations))
	}

	if len(listResult.Organizations) > 0 && listResult.Organizations[0].ID != org1Result.Organization.ID {
		t.Errorf("ORG_ADMIN should only see their own org %s, got %s", 
			org1Result.Organization.ID, listResult.Organizations[0].ID)
	}

	// Try to access Hospital B (should be forbidden)
	getOrg2Resp := org1AdminClient.GET(t, "/organizations/"+org2Result.Organization.ID)
	if getOrg2Resp.StatusCode != http.StatusForbidden {
		t.Errorf("ORG_ADMIN should not access other org, expected 403, got %d", getOrg2Resp.StatusCode)
	}

	t.Logf("E2E Test Passed: ORG_ADMIN correctly isolated to their own organization")
}

// TestE2E_MultiTenant_SuperAdminSeesAllOrgs tests that SUPER_ADMIN can see all organizations
func TestE2E_MultiTenant_SuperAdminSeesAllOrgs(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create multiple organizations
	orgCount := 3
	createdOrgIDs := make([]string, orgCount)

	for i := 0; i < orgCount; i++ {
		orgBody := map[string]interface{}{
			"name":          "Multi-Tenant Hospital " + string(rune('A'+i)),
			"contact_email": "mt" + string(rune('a'+i)) + "@test.com",
		}

		orgResp := client.POST(t, "/organizations", orgBody)
		testutil.AssertStatusCode(t, orgResp, http.StatusCreated)

		var orgResult struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		}
		testutil.DecodeJSON(t, orgResp, &orgResult)
		createdOrgIDs[i] = orgResult.Organization.ID
	}

	// SUPER_ADMIN lists all organizations
	listResp := client.GET(t, "/organizations?page=1&limit=100")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Organizations []struct {
			ID string `json:"id"`
		} `json:"organizations"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// Should see at least the organizations we created
	if len(listResult.Organizations) < orgCount {
		t.Errorf("SUPER_ADMIN should see at least %d organizations, got %d", 
			orgCount, len(listResult.Organizations))
	}

	// Verify all created orgs are in the list
	foundCount := 0
	for _, org := range listResult.Organizations {
		for _, createdID := range createdOrgIDs {
			if org.ID == createdID {
				foundCount++
			}
		}
	}

	if foundCount != orgCount {
		t.Errorf("SUPER_ADMIN should see all %d created organizations, found %d", orgCount, foundCount)
	}

	t.Logf("E2E Test Passed: SUPER_ADMIN can see all %d organizations", len(listResult.Organizations))
}

// TestE2E_MultiTenant_SchemaIsolation tests that tenant schemas are properly isolated
func TestE2E_MultiTenant_SchemaIsolation(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	client := ts.NewClient(superAdminToken)

	// Create two organizations with different schemas
	org1Body := map[string]interface{}{
		"name":          "Schema Test Hospital 1",
		"contact_email": "schema1@test.com",
	}
	org1Resp := client.POST(t, "/organizations", org1Body)
	testutil.AssertStatusCode(t, org1Resp, http.StatusCreated)

	var org1Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org1Resp, &org1Result)

	org2Body := map[string]interface{}{
		"name":          "Schema Test Hospital 2",
		"contact_email": "schema2@test.com",
	}
	org2Resp := client.POST(t, "/organizations", org2Body)
	testutil.AssertStatusCode(t, org2Resp, http.StatusCreated)

	var org2Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org2Resp, &org2Result)

	// Verify both schemas exist in database
	var schema1Exists, schema2Exists bool

	err := ts.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, org1Result.Organization.SchemaName).Scan(&schema1Exists)

	if err != nil {
		t.Fatalf("Failed to check schema1 existence: %v", err)
	}

	err = ts.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, org2Result.Organization.SchemaName).Scan(&schema2Exists)

	if err != nil {
		t.Fatalf("Failed to check schema2 existence: %v", err)
	}

	if !schema1Exists {
		t.Errorf("Schema %s should exist", org1Result.Organization.SchemaName)
	}

	if !schema2Exists {
		t.Errorf("Schema %s should exist", org2Result.Organization.SchemaName)
	}

	// Verify schemas are different
	if org1Result.Organization.SchemaName == org2Result.Organization.SchemaName {
		t.Error("Different organizations should have different schemas")
	}

	// Verify each schema has the required tables (users, patients, etc.)
	for _, schemaName := range []string{org1Result.Organization.SchemaName, org2Result.Organization.SchemaName} {
		var tableCount int
		err := ts.DB.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = $1
		`, schemaName).Scan(&tableCount)

		if err != nil {
			t.Fatalf("Failed to count tables in schema %s: %v", schemaName, err)
		}

		if tableCount == 0 {
			t.Errorf("Schema %s should have tables, got 0", schemaName)
		}
	}

	t.Logf("E2E Test Passed: Tenant schemas are properly isolated (%s vs %s)", 
		org1Result.Organization.SchemaName, org2Result.Organization.SchemaName)
}

// TestE2E_MultiTenant_UserCannotAccessOtherOrgData tests that users from one org cannot access another org's data
func TestE2E_MultiTenant_UserCannotAccessOtherOrgData(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	superAdminClient := ts.NewClient(superAdminToken)

	// Create two organizations
	org1Body := map[string]interface{}{
		"name":          "Hospital A for User Isolation",
		"contact_email": "hospitala@isolation.com",
	}
	org1Resp := superAdminClient.POST(t, "/organizations", org1Body)
	testutil.AssertStatusCode(t, org1Resp, http.StatusCreated)

	var org1Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org1Resp, &org1Result)

	org2Body := map[string]interface{}{
		"name":          "Hospital B for User Isolation",
		"contact_email": "hospitalb@isolation.com",
	}
	org2Resp := superAdminClient.POST(t, "/organizations", org2Body)
	testutil.AssertStatusCode(t, org2Resp, http.StatusCreated)

	var org2Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org2Resp, &org2Result)

	// Create user in Hospital A
	userBody := map[string]interface{}{
		"username":          "usera",
		"email":             "usera@hospitala.com",
		"firstName":         "User",
		"lastName":          "A",
		"role":              "CAREGIVER",
		"temporaryPassword": "TempPass123!",
	}

	userResp := superAdminClient.POSTWithOrgHeader(t, "/organization/users", userBody, org1Result.Organization.ID)
	testutil.AssertStatusCode(t, userResp, http.StatusCreated)

	var userResult struct {
		ID string `json:"id"`
	}
	testutil.DecodeJSON(t, userResp, &userResult)
	userAID := userResult.ID

	// ORG_ADMIN from Hospital B tries to access user from Hospital A (should be forbidden)
	org2AdminToken := ts.GenerateOrgAdminToken(t, org2Result.Organization.ID, org2Result.Organization.SchemaName)
	org2AdminClient := ts.NewClient(org2AdminToken)

	// Try to list users from Hospital A (should only see Hospital B users, which is none)
	listResp := org2AdminClient.GETWithOrgHeader(t, "/organization/users", org1Result.Organization.ID)
	
	// Should be forbidden (403) when trying to access different org
	if listResp.StatusCode != http.StatusForbidden {
		// Or might return empty list depending on implementation
		// Let's check if it returns empty
		var listResult struct {
			Users []struct {
				ID string `json:"id"`
			} `json:"users"`
		}
		testutil.DecodeJSON(t, listResp, &listResult)
		
		// Should not see user from Hospital A
		for _, user := range listResult.Users {
			if user.ID == userAID {
				t.Error("ORG_ADMIN from Hospital B should not see users from Hospital A")
			}
		}
	}

	t.Logf("E2E Test Passed: Multi-tenant user data isolation verified")
}

// TestE2E_MultiTenant_PatientDataIsolation tests that patients from one org cannot be accessed by another org
func TestE2E_MultiTenant_PatientDataIsolation(t *testing.T) {
	ts := SetupE2ETest(t)
	defer ts.Cleanup(t)

	superAdminToken := ts.GenerateSuperAdminToken(t)
	superAdminClient := ts.NewClient(superAdminToken)

	// Create two organizations
	org1Body := map[string]interface{}{
		"name":          "Hospital A for Patient Isolation",
		"contact_email": "hospitala@patisolation.com",
	}
	org1Resp := superAdminClient.POST(t, "/organizations", org1Body)
	testutil.AssertStatusCode(t, org1Resp, http.StatusCreated)

	var org1Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org1Resp, &org1Result)

	org2Body := map[string]interface{}{
		"name":          "Hospital B for Patient Isolation",
		"contact_email": "hospitalb@patisolation.com",
	}
	org2Resp := superAdminClient.POST(t, "/organizations", org2Body)
	testutil.AssertStatusCode(t, org2Resp, http.StatusCreated)

	var org2Result struct {
		Organization struct {
			ID         string `json:"id"`
			SchemaName string `json:"schema_name"`
		} `json:"organization"`
	}
	testutil.DecodeJSON(t, org2Resp, &org2Result)

	// Create patient in Hospital A
	patientBody := map[string]interface{}{
		"username":          "patienta",
		"email":             "patienta@hospitala.com",
		"firstName":         "Patient",
		"lastName":          "A",
		"dateOfBirth":       "1980-01-01",
		"address":           "123 Hospital A St",
		"temporaryPassword": "TempPass123!",
	}

	patientResp := superAdminClient.POSTWithOrgHeader(t, "/organization/patients", patientBody, org1Result.Organization.ID)
	testutil.AssertStatusCode(t, patientResp, http.StatusCreated)

	var patientResult struct {
		Patient struct {
			ID string `json:"id"`
		} `json:"patient"`
	}
	testutil.DecodeJSON(t, patientResp, &patientResult)
	patientAID := patientResult.Patient.ID

	// ORG_ADMIN from Hospital B tries to access patients from Hospital A
	org2AdminToken := ts.GenerateOrgAdminToken(t, org2Result.Organization.ID, org2Result.Organization.SchemaName)
	org2AdminClient := ts.NewClient(org2AdminToken)

	// Try to list patients (should only see Hospital B patients, which is none)
	listResp := org2AdminClient.GET(t, "/organization/patients")
	testutil.AssertStatusCode(t, listResp, http.StatusOK)

	var listResult struct {
		Patients []struct {
			ID string `json:"id"`
		} `json:"patients"`
	}
	testutil.DecodeJSON(t, listResp, &listResult)

	// Should not see patient from Hospital A
	for _, patient := range listResult.Patients {
		if patient.ID == patientAID {
			t.Error("ORG_ADMIN from Hospital B should not see patients from Hospital A")
		}
	}

	// Try to directly access patient from Hospital A (should fail or return 404)
	getResp := org2AdminClient.GET(t, "/organization/patients/"+patientAID)
	
	// Should not be able to access (404 because it's in different schema)
	if getResp.StatusCode == http.StatusOK {
		t.Error("ORG_ADMIN from Hospital B should not be able to access patient from Hospital A")
	}

	t.Logf("E2E Test Passed: Multi-tenant patient data isolation verified")
}
