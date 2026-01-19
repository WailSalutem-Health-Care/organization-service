package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds all custom metrics for the service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    metric.Int64Counter
	HTTPDurationMs       metric.Float64Histogram

	// Business metrics
	OrganizationTotal    metric.Int64Counter
	PatientTotal         metric.Int64Counter
	UserTotal            metric.Int64Counter

	// Auth metrics
	AuthFailuresTotal    metric.Int64Counter
	PermissionCheckDuration metric.Float64Histogram
}

// InitMetrics initializes all custom metrics
func InitMetrics() (*Metrics, error) {
	meter := otel.Meter("github.com/WailSalutem-Health-Care/organization-service")

	// HTTP request counter
	httpRequestsTotal, err := meter.Int64Counter(
		"http_server_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	// HTTP duration histogram
	httpDurationMs, err := meter.Float64Histogram(
		"http_server_duration_milliseconds",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	// Organization counter
	organizationTotal, err := meter.Int64Counter(
		"organization_total",
		metric.WithDescription("Total number of organization operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	// Patient counter
	patientTotal, err := meter.Int64Counter(
		"patient_total",
		metric.WithDescription("Total number of patient operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	// User counter
	userTotal, err := meter.Int64Counter(
		"user_total",
		metric.WithDescription("Total number of user operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	// Auth failures counter
	authFailuresTotal, err := meter.Int64Counter(
		"auth_failures_total",
		metric.WithDescription("Total number of authentication failures"),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return nil, err
	}

	// Permission check duration histogram
	permissionCheckDuration, err := meter.Float64Histogram(
		"permission_check_duration_ms",
		metric.WithDescription("Permission check duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	log.Println("✓ Custom metrics initialized")

	return &Metrics{
		HTTPRequestsTotal:       httpRequestsTotal,
		HTTPDurationMs:          httpDurationMs,
		OrganizationTotal:       organizationTotal,
		PatientTotal:            patientTotal,
		UserTotal:               userTotal,
		AuthFailuresTotal:       authFailuresTotal,
		PermissionCheckDuration: permissionCheckDuration,
	}, nil
}

// RecordHTTPRequest records an HTTP request metric
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, durationMs float64) {
	attrs := []attribute.KeyValue{
		attribute.String("http_method", method),
		attribute.String("http_route", route),
		attribute.Int("http_status_code", statusCode),
	}

	m.HTTPRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.HTTPDurationMs.Record(ctx, durationMs, metric.WithAttributes(attrs...))
}

// RecordOrganizationOperation records an organization operation metric
func (m *Metrics) RecordOrganizationOperation(ctx context.Context, operation string) {
	m.OrganizationTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
	))
}

// RecordPatientOperation records a patient operation metric
func (m *Metrics) RecordPatientOperation(ctx context.Context, operation string) {
	m.PatientTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
	))
}

// RecordUserOperation records a user operation metric
func (m *Metrics) RecordUserOperation(ctx context.Context, operation string) {
	m.UserTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
	))
}

// RecordAuthFailure records an authentication failure metric
func (m *Metrics) RecordAuthFailure(ctx context.Context, reason string) {
	m.AuthFailuresTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("reason", reason),
	))
}

// RecordPermissionCheck records a permission check duration metric
func (m *Metrics) RecordPermissionCheck(ctx context.Context, permission string, durationMs float64, allowed bool) {
	m.PermissionCheckDuration.Record(ctx, durationMs, metric.WithAttributes(
		attribute.String("permission", permission),
		attribute.Bool("allowed", allowed),
	))
}
