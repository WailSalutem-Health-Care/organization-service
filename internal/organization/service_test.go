package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// TestCreateOrganization_Success tests successful organization creation
func TestCreateOrganization_Success(t *testing.T) {
	mockRepo := &mockRepository{
		createOrgFunc: func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:           "org-123",
				Name:         req.Name,
				SchemaName:   "org_testorg_12345678",
				ContactEmail: req.ContactEmail,
				ContactPhone: req.ContactPhone,
				Address:      req.Address,
				Status:       "active",
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	service := NewService(mockRepo)
	req := CreateOrganizationRequest{
		Name:         "Test Org",
		ContactEmail: "test@example.com",
		ContactPhone: "+1234567890",
		Address:      "123 Test St",
	}

	org, err := service.CreateOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if org == nil {
		t.Fatal("Expected organization, got nil")
	}
	if org.Name != "Test Org" {
		t.Errorf("Expected name 'Test Org', got '%s'", org.Name)
	}
	if org.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", org.Status)
	}
}

// TestCreateOrganization_EmptyName tests validation for empty name
func TestCreateOrganization_EmptyName(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	req := CreateOrganizationRequest{
		Name:         "",
		ContactEmail: "test@example.com",
	}

	org, err := service.CreateOrganization(context.Background(), req)

	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
	if org != nil {
		t.Error("Expected nil organization")
	}
	if err.Error() != "organization name is required" {
		t.Errorf("Expected 'organization name is required', got '%s'", err.Error())
	}
}

// TestCreateOrganization_RepositoryError tests handling of repository errors
func TestCreateOrganization_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		createOrgFunc: func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}

	service := NewService(mockRepo)
	req := CreateOrganizationRequest{
		Name: "Test Org",
	}

	org, err := service.CreateOrganization(context.Background(), req)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if org != nil {
		t.Error("Expected nil organization")
	}
}

// TestListOrganizations_SuperAdmin tests SUPER_ADMIN sees all organizations
func TestListOrganizations_SuperAdmin(t *testing.T) {
	mockRepo := &mockRepository{
		listOrgsFunc: func(ctx context.Context) ([]OrganizationResponse, error) {
			return []OrganizationResponse{
				{ID: "org-1", Name: "Org 1", Status: "active"},
				{ID: "org-2", Name: "Org 2", Status: "active"},
				{ID: "org-3", Name: "Org 3", Status: "active"},
			}, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "org-1",
	}

	orgs, err := service.ListOrganizations(context.Background(), principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(orgs) != 3 {
		t.Errorf("Expected 3 organizations, got %d", len(orgs))
	}
}

// TestListOrganizations_OrgAdmin tests ORG_ADMIN sees only their organization
func TestListOrganizations_OrgAdmin(t *testing.T) {
	mockRepo := &mockRepository{
		getOrgFunc: func(ctx context.Context, id string) (*OrganizationResponse, error) {
			if id == "org-2" {
				return &OrganizationResponse{
					ID:     "org-2",
					Name:   "Org 2",
					Status: "active",
				}, nil
			}
			return nil, errors.New("organization not found")
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-2",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-2",
	}

	orgs, err := service.ListOrganizations(context.Background(), principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(orgs) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(orgs))
	}
	if orgs[0].ID != "org-2" {
		t.Errorf("Expected org-2, got %s", orgs[0].ID)
	}
}

// TestListOrganizations_OrgAdminNoOrgID tests error when ORG_ADMIN has no org_id
func TestListOrganizations_OrgAdminNoOrgID(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	principal := &auth.Principal{
		UserID: "user-3",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "", // No org ID
	}

	orgs, err := service.ListOrganizations(context.Background(), principal)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if orgs != nil {
		t.Error("Expected nil organizations")
	}
	if err.Error() != "no organization associated with this user" {
		t.Errorf("Expected 'no organization associated', got '%s'", err.Error())
	}
}

// TestListOrganizationsWithPagination_SuperAdmin tests pagination for SUPER_ADMIN
func TestListOrganizationsWithPagination_SuperAdmin(t *testing.T) {
	mockRepo := &mockRepository{
		listOrgsPaginatedFunc: func(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error) {
			// Simulate 25 total orgs, returning page of 10
			return []OrganizationResponse{
				{ID: "org-1", Name: "Org 1"},
				{ID: "org-2", Name: "Org 2"},
				{ID: "org-3", Name: "Org 3"},
				{ID: "org-4", Name: "Org 4"},
				{ID: "org-5", Name: "Org 5"},
				{ID: "org-6", Name: "Org 6"},
				{ID: "org-7", Name: "Org 7"},
				{ID: "org-8", Name: "Org 8"},
				{ID: "org-9", Name: "Org 9"},
				{ID: "org-10", Name: "Org 10"},
			}, 25, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
	}

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListOrganizationsWithPagination(context.Background(), principal, params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !response.Success {
		t.Error("Expected success to be true")
	}
	if len(response.Organizations) != 10 {
		t.Errorf("Expected 10 organizations, got %d", len(response.Organizations))
	}
	if response.Pagination.TotalRecords != 25 {
		t.Errorf("Expected total count 25, got %d", response.Pagination.TotalRecords)
	}
	if response.Pagination.TotalPages != 3 {
		t.Errorf("Expected 3 pages, got %d", response.Pagination.TotalPages)
	}
}

// TestListOrganizationsWithPagination_WithSearch tests search functionality
func TestListOrganizationsWithPagination_WithSearch(t *testing.T) {
	mockRepo := &mockRepository{
		listOrgsPaginatedFunc: func(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error) {
			if search == "hospital" {
				return []OrganizationResponse{
					{ID: "org-1", Name: "City Hospital"},
					{ID: "org-5", Name: "General Hospital"},
				}, 2, nil
			}
			return []OrganizationResponse{}, 0, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
	}

	params := pagination.Params{
		Page:   1,
		Limit:  10,
		Search: "hospital",
	}

	response, err := service.ListOrganizationsWithPagination(context.Background(), principal, params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(response.Organizations))
	}
}

// TestListOrganizationsWithPagination_WithStatusFilter tests status filtering
func TestListOrganizationsWithPagination_WithStatusFilter(t *testing.T) {
	mockRepo := &mockRepository{
		listOrgsPaginatedFunc: func(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error) {
			if status == "active" {
				return []OrganizationResponse{
					{ID: "org-1", Name: "Active Org 1", Status: "active"},
					{ID: "org-2", Name: "Active Org 2", Status: "active"},
				}, 2, nil
			}
			return []OrganizationResponse{}, 0, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
	}

	params := pagination.Params{
		Page:   1,
		Limit:  10,
		Status: "active",
	}

	response, err := service.ListOrganizationsWithPagination(context.Background(), principal, params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Organizations) != 2 {
		t.Errorf("Expected 2 active organizations, got %d", len(response.Organizations))
	}
}

// TestListOrganizationsWithPagination_OrgAdmin tests ORG_ADMIN only sees their org
func TestListOrganizationsWithPagination_OrgAdmin(t *testing.T) {
	mockRepo := &mockRepository{
		getOrgFunc: func(ctx context.Context, id string) (*OrganizationResponse, error) {
			if id == "org-3" {
				return &OrganizationResponse{
					ID:     "org-3",
					Name:   "My Org",
					Status: "active",
				}, nil
			}
			return nil, errors.New("organization not found")
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-3",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-3",
	}

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListOrganizationsWithPagination(context.Background(), principal, params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(response.Organizations))
	}
	if response.Pagination.TotalRecords != 1 {
		t.Errorf("Expected total count 1, got %d", response.Pagination.TotalRecords)
	}
}

// TestGetOrganization_SuperAdmin tests SUPER_ADMIN can view any organization
func TestGetOrganization_SuperAdmin(t *testing.T) {
	mockRepo := &mockRepository{
		getOrgFunc: func(ctx context.Context, id string) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:     id,
				Name:   "Test Org",
				Status: "active",
			}, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "org-1",
	}

	org, err := service.GetOrganization(context.Background(), "org-999", principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if org.ID != "org-999" {
		t.Errorf("Expected org-999, got %s", org.ID)
	}
}

// TestGetOrganization_OrgAdminOwnOrg tests ORG_ADMIN viewing their own org
func TestGetOrganization_OrgAdminOwnOrg(t *testing.T) {
	mockRepo := &mockRepository{
		getOrgFunc: func(ctx context.Context, id string) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:     id,
				Name:   "My Org",
				Status: "active",
			}, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-2",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-2",
	}

	org, err := service.GetOrganization(context.Background(), "org-2", principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if org.ID != "org-2" {
		t.Errorf("Expected org-2, got %s", org.ID)
	}
}

// TestGetOrganization_OrgAdminForbidden tests ORG_ADMIN cannot view other orgs
func TestGetOrganization_OrgAdminForbidden(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	principal := &auth.Principal{
		UserID: "user-2",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-2",
	}

	// Trying to view org-3 when user belongs to org-2
	org, err := service.GetOrganization(context.Background(), "org-3", principal)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if org != nil {
		t.Error("Expected nil organization")
	}
	if err.Error() != "forbidden" {
		t.Errorf("Expected 'forbidden', got '%s'", err.Error())
	}
}

// TestUpdateOrganization_SuperAdmin tests SUPER_ADMIN can update any organization
func TestUpdateOrganization_SuperAdmin(t *testing.T) {
	newName := "Updated Org"
	mockRepo := &mockRepository{
		updateOrgFunc: func(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:     id,
				Name:   *req.Name,
				Status: "active",
			}, nil
		},
	}

	service := NewService(mockRepo)
	principal := &auth.Principal{
		UserID: "user-1",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "org-1",
	}

	req := UpdateOrganizationRequest{
		Name: &newName,
	}

	org, err := service.UpdateOrganization(context.Background(), "org-5", req, principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if org.Name != "Updated Org" {
		t.Errorf("Expected 'Updated Org', got '%s'", org.Name)
	}
}

// TestUpdateOrganization_OrgAdminForbidden tests ORG_ADMIN cannot update
func TestUpdateOrganization_OrgAdminForbidden(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)

	principal := &auth.Principal{
		UserID: "user-2",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-2",
	}

	newName := "Updated Org"
	req := UpdateOrganizationRequest{
		Name: &newName,
	}

	org, err := service.UpdateOrganization(context.Background(), "org-2", req, principal)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if org != nil {
		t.Error("Expected nil organization")
	}
	if err.Error() != "forbidden" {
		t.Errorf("Expected 'forbidden', got '%s'", err.Error())
	}
}

// TestDeleteOrganization_Success tests successful organization deletion
func TestDeleteOrganization_Success(t *testing.T) {
	mockRepo := &mockRepository{
		deleteOrgFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	service := NewService(mockRepo)
	err := service.DeleteOrganization(context.Background(), "org-1")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestDeleteOrganization_NotFound tests deleting non-existent organization
func TestDeleteOrganization_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		deleteOrgFunc: func(ctx context.Context, id string) error {
			return errors.New("organization not found or already deleted")
		},
	}

	service := NewService(mockRepo)
	err := service.DeleteOrganization(context.Background(), "nonexistent")

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Mock repository for testing
type mockRepository struct {
	createOrgFunc         func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error)
	listOrgsFunc          func(ctx context.Context) ([]OrganizationResponse, error)
	listOrgsPaginatedFunc func(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error)
	getOrgFunc            func(ctx context.Context, id string) (*OrganizationResponse, error)
	updateOrgFunc         func(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error)
	deleteOrgFunc         func(ctx context.Context, id string) error
}

func (m *mockRepository) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	if m.createOrgFunc != nil {
		return m.createOrgFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) ListOrganizations(ctx context.Context) ([]OrganizationResponse, error) {
	if m.listOrgsFunc != nil {
		return m.listOrgsFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) ListOrganizationsWithPagination(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error) {
	if m.listOrgsPaginatedFunc != nil {
		return m.listOrgsPaginatedFunc(ctx, limit, offset, search, status)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepository) GetOrganization(ctx context.Context, id string) (*OrganizationResponse, error) {
	if m.getOrgFunc != nil {
		return m.getOrgFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
	if m.updateOrgFunc != nil {
		return m.updateOrgFunc(ctx, id, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) DeleteOrganization(ctx context.Context, id string) error {
	if m.deleteOrgFunc != nil {
		return m.deleteOrgFunc(ctx, id)
	}
	return errors.New("not implemented")
}
