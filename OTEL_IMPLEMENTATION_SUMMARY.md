# OpenTelemetry Implementation Summary

## Overview

This document summarizes the OpenTelemetry instrumentation implementation for the organization-service. All tasks have been completed successfully, and the service now has comprehensive observability capabilities.

## ✅ Completed Tasks

### 1. Core Telemetry Infrastructure

**Created**: `internal/telemetry/` package

- **`telemetry.go`**: Core initialization logic
  - Tracer provider with OTLP gRPC exporter
  - Meter provider for metrics
  - Resource attributes (service name, namespace, version, environment)
  - Graceful shutdown with timeout
  - Fails gracefully if OTLP collector is unavailable

- **`metrics.go`**: Custom metrics definitions
  - HTTP server metrics (requests, duration)
  - Business metrics (organization, patient, user operations)
  - Auth metrics (failures, permission checks)
  - Helper functions for recording metrics

### 2. Dependencies Updated

**Modified**: `go.mod`

Added OpenTelemetry packages:
- `go.opentelemetry.io/otel v1.39.0`
- `go.opentelemetry.io/otel/sdk v1.39.0`
- `go.opentelemetry.io/otel/metric v1.39.0`
- `go.opentelemetry.io/otel/sdk/metric v1.39.0`
- `go.opentelemetry.io/otel/trace v1.39.0`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0`
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.44.0`
- `go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.64.0`
- `github.com/XSAM/otelsql v0.41.0`
- `google.golang.org/grpc v1.60.1`

### 3. Main Application Instrumentation

**Modified**: `cmd/api/main.go`

Changes:
- Initialize OpenTelemetry providers on startup
- Initialize custom metrics
- Wrap Gorilla Mux router with `otelmux.Middleware`
- Add graceful shutdown for telemetry providers
- Add signal handling for clean shutdown
- HTTP server with proper timeouts

Key features:
- Service continues if OTLP collector is unavailable
- Telemetry is flushed on graceful shutdown
- All traces include service resource attributes

### 4. HTTP Router Updates

**Modified**: `internal/http/router.go`

Changes:
- Added metrics parameter to `SetupRouter` and `SetupRouterWithKeycloak`
- Updated all route handlers to use metrics-aware middleware:
  - `auth.MiddlewareWithMetrics` for authentication
  - `auth.RequirePermissionWithMetrics` for authorization
- Metrics are passed to all middleware layers

### 5. PostgreSQL Instrumentation

**Modified**: `internal/db/postgres.go`

Changes:
- Wrapped `sql.Open` with `otelsql.Open`
- Added database resource attributes (system, name)
- Registered DB stats metrics for connection pool monitoring
- Automatic query tracing with duration and row counts

Captured metrics:
- Active/idle connections
- Connection wait time
- Query duration histograms

### 6. RabbitMQ Instrumentation

**Modified**: `internal/messaging/rabbitmq.go`

Changes:
- Added tracer initialization
- Wrapped `Publish` method with span creation
- Implemented trace context propagation via message headers
- Created `rabbitMQCarrier` for TextMapPropagator interface
- Added span attributes (exchange, routing key, message size)

Features:
- Trace context flows through message queue
- Producer spans with proper attributes
- Error recording in spans

### 7. Keycloak HTTP Client Instrumentation

**Modified**: `internal/auth/keycloak_admin.go`

Changes:
- Created custom `otelTransport` wrapper
- Wrapped HTTP client transport with OpenTelemetry
- Automatic tracing of all Keycloak API calls
- Captures HTTP method, URL, status code, duration

Features:
- External API call visibility
- Error tracking
- Request/response correlation

### 8. Authentication & Authorization Instrumentation

**Modified**: `internal/auth/middleware.go`

Changes:
- Added tracer initialization
- Created `MetricsRecorder` and `PermissionMetricsRecorder` interfaces
- Implemented `MiddlewareWithMetrics` for token validation
- Implemented `RequirePermissionWithMetrics` for permission checks
- Added span attributes (user ID, roles, organization ID, permissions)

Captured data:
- Authentication failures with reasons
- Permission check duration
- User context in all downstream spans

### 9. Environment Configuration

**Created**: `.env.example`

Added OpenTelemetry environment variables:
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_SERVICE_NAME=organization-service
OTEL_SERVICE_NAMESPACE=wailsalutem
OTEL_SERVICE_VERSION=1.0.0
ENVIRONMENT=production
OTEL_TRACES_SAMPLER=always_on
OTEL_METRICS_EXPORT_INTERVAL=30s
```

### 10. Docker Compose Updates

**Modified**: `docker-compose.yml`

Changes:
- Added all OTEL environment variables to organization-service
- Environment variables with defaults using `${VAR:-default}` syntax
- Service can connect to OTLP collector in same network

### 11. Comprehensive Documentation

**Created**: `OBSERVABILITY.md`

Complete guide including:
- Architecture diagrams
- Configuration reference
- Instrumentation details for each layer
- Local testing setup with full observability stack
- Docker Compose configurations for:
  - OpenTelemetry Collector
  - Tempo (traces)
  - Prometheus (metrics)
  - Grafana (visualization)
- Metrics reference with all custom metrics
- Traces reference with span structure
- Prometheus query examples
- Tempo query examples
- Grafana dashboard recommendations
- Troubleshooting guide
- Best practices

---

## Architecture Summary

### Telemetry Flow

```
HTTP Request
    ↓
[otelmux Middleware] ← HTTP span created
    ↓
[auth.MiddlewareWithMetrics] ← Auth span + metrics
    ↓
[auth.RequirePermissionWithMetrics] ← Permission span + metrics
    ↓
[Handler] → [Service] → [Repository]
                ↓           ↓
            [RabbitMQ]  [PostgreSQL]
                ↓           ↓
            Producer    Query spans
             spans      (automatic)
                ↓
          [Keycloak]
              ↓
         HTTP client
            spans
```

### Instrumented Components

1. **HTTP Layer**
   - ✅ Gorilla Mux automatic instrumentation
   - ✅ Request/response tracing
   - ✅ Route-level metrics

2. **Authentication Layer**
   - ✅ Token validation spans
   - ✅ Permission check spans
   - ✅ Auth failure metrics
   - ✅ User context propagation

3. **Database Layer**
   - ✅ Automatic SQL query tracing
   - ✅ Connection pool metrics
   - ✅ Query duration histograms

4. **Messaging Layer**
   - ✅ RabbitMQ publish spans
   - ✅ Trace context propagation
   - ✅ Message attributes

5. **External Services**
   - ✅ Keycloak HTTP client tracing
   - ✅ Request/response correlation

---

## Metrics Catalog

### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_server_requests_total` | Counter | `http_method`, `http_route`, `http_status_code` | Total HTTP requests |
| `http_server_duration_milliseconds` | Histogram | `http_method`, `http_route`, `http_status_code` | Request duration |

### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `organization_total` | Counter | `operation` | Organization operations |
| `patient_total` | Counter | `operation` | Patient operations |
| `user_total` | Counter | `operation` | User operations |

### Auth Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `auth_failures_total` | Counter | `reason` | Authentication failures |
| `permission_check_duration_ms` | Histogram | `permission`, `allowed` | Permission check latency |

### Database Metrics (Automatic)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `db.client.connections.usage` | Gauge | `db.system`, `db.name`, `state` | Connection pool usage |
| `db.client.connections.max` | Gauge | `db.system`, `db.name` | Max connections |
| `db.client.connections.wait_time` | Histogram | `db.system`, `db.name` | Connection wait time |

---

## Trace Attributes

### Standard Attributes

Every span includes:
- `service.name`: organization-service
- `service.namespace`: wailsalutem
- `service.version`: 1.0.0
- `deployment.environment`: production

### HTTP Spans

- `http.method`: GET, POST, PUT, DELETE
- `http.route`: /organizations/{id}
- `http.status_code`: 200, 201, 400, 500
- `http.target`: Full request path

### Database Spans

- `db.system`: postgresql
- `db.name`: organization_db
- `db.statement`: SQL query (sanitized)
- `db.operation`: SELECT, INSERT, UPDATE, DELETE

### Messaging Spans

- `messaging.system`: rabbitmq
- `messaging.destination`: wailsalutem.events
- `messaging.routing_key`: patient.created
- `messaging.message_payload_size_bytes`: 256

### User Context

- `user.id`: User UUID
- `user.email`: User email
- `user.roles`: User roles array
- `organization.id`: Organization UUID

---

## Quick Start

### 1. Set Environment Variables

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_SERVICE_NAME=organization-service
export OTEL_SERVICE_NAMESPACE=wailsalutem
export OTEL_SERVICE_VERSION=1.0.0
export ENVIRONMENT=development
export OTEL_TRACES_SAMPLER=always_on
```

### 2. Start Observability Stack

See `OBSERVABILITY.md` for complete Docker Compose setup with:
- OpenTelemetry Collector
- Tempo (traces)
- Prometheus (metrics)
- Grafana (dashboards)

### 3. Start the Service

```bash
go run cmd/api/main.go
```

Or with Docker:

```bash
docker-compose up -d
```

### 4. Access Dashboards

- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200

---

## Testing the Implementation

### 1. Verify Telemetry Initialization

Check logs on startup:
```
✓ OpenTelemetry tracer provider initialized
✓ OpenTelemetry meter provider initialized
✓ Custom metrics initialized
```

### 2. Generate Test Traffic

```bash
# Create organization
curl -X POST http://localhost:8080/organizations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Org"}'

# List organizations
curl http://localhost:8080/organizations \
  -H "Authorization: Bearer $TOKEN"
```

### 3. View Traces in Grafana

1. Open Grafana: http://localhost:3000
2. Go to Explore
3. Select Tempo data source
4. Search for traces: `{service.name="organization-service"}`

### 4. View Metrics in Prometheus

1. Open Prometheus: http://localhost:9090
2. Query metrics:
   ```promql
   rate(http_server_requests_total[5m])
   histogram_quantile(0.95, rate(http_server_duration_milliseconds_bucket[5m]))
   organization_total
   ```

---

## Key Features

### ✅ Comprehensive Coverage

- HTTP server (Gorilla Mux)
- PostgreSQL database
- RabbitMQ messaging
- Keycloak HTTP client
- Authentication & authorization
- Custom business metrics

### ✅ Production-Ready

- Graceful degradation (service continues if collector unavailable)
- Configurable sampling
- Batch processing for efficiency
- Resource attributes for multi-service environments
- Proper context propagation

### ✅ Developer-Friendly

- Clear environment variable configuration
- Comprehensive documentation
- Example queries and dashboards
- Troubleshooting guide
- Local testing setup

### ✅ Best Practices

- Semantic conventions (OpenTelemetry standard)
- Proper span naming
- Error recording
- Low-cardinality labels
- Context propagation through all layers

---

## Next Steps (Optional Enhancements)

### 1. Add Service-Level Business Spans

To add custom spans in service methods:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/WailSalutem-Health-Care/organization-service/organization")

func (s *Service) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
    ctx, span := tracer.Start(ctx, "organization.CreateOrganization")
    defer span.End()
    
    span.SetAttributes(attribute.String("organization.name", req.Name))
    
    // ... business logic ...
    
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    
    return org, nil
}
```

### 2. Add Custom Dashboards

Create Grafana dashboards for:
- Service overview (requests, errors, latency)
- Business metrics (operations per minute)
- Database performance
- Auth metrics

### 3. Set Up Alerts

Configure Prometheus alerts for:
- High error rate
- Slow response times
- Database connection pool exhaustion
- Auth failure spikes

### 4. Add Logs Correlation

Integrate structured logging with trace IDs:
```go
import "go.opentelemetry.io/otel/trace"

traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
log.Printf("[trace_id=%s] Processing request", traceID)
```

---

## Troubleshooting

### Service Won't Start

**Symptom**: Service fails with OTLP connection error

**Solution**: This is expected if the collector isn't running. The service will log a warning and continue:
```
Warning: failed to initialize OpenTelemetry: ...
Service will continue without observability
```

### No Traces Appearing

**Check**:
1. OTLP collector is running: `curl http://localhost:13133`
2. Environment variable: `echo $OTEL_EXPORTER_OTLP_ENDPOINT`
3. Sampling: `OTEL_TRACES_SAMPLER=always_on`

### Build Verification

The implementation has been tested and builds successfully:
```bash
$ go build -o /tmp/org-service-test ./cmd/api/main.go
# Success - no errors
```

---

## Files Modified/Created

### Created Files
- `internal/telemetry/telemetry.go` - Core telemetry initialization
- `internal/telemetry/metrics.go` - Custom metrics definitions
- `OBSERVABILITY.md` - Comprehensive observability guide
- `OTEL_IMPLEMENTATION_SUMMARY.md` - This file

### Modified Files
- `go.mod` - Added OpenTelemetry dependencies
- `go.sum` - Updated dependency checksums
- `cmd/api/main.go` - Initialize telemetry, add graceful shutdown
- `internal/http/router.go` - Pass metrics to middleware
- `internal/db/postgres.go` - Add PostgreSQL instrumentation
- `internal/messaging/rabbitmq.go` - Add RabbitMQ instrumentation
- `internal/auth/keycloak_admin.go` - Add HTTP client instrumentation
- `internal/auth/middleware.go` - Add auth/authz instrumentation
- `.env.example` - Add OTEL environment variables
- `docker-compose.yml` - Add OTEL configuration

---

## Summary

The organization-service now has **production-ready OpenTelemetry instrumentation** with:

- ✅ Distributed tracing across all components
- ✅ Custom business and technical metrics
- ✅ Automatic database and HTTP instrumentation
- ✅ Trace context propagation through message queues
- ✅ External service call tracking
- ✅ Authentication and authorization observability
- ✅ Graceful degradation
- ✅ Comprehensive documentation
- ✅ Local testing setup
- ✅ Build verification passed

The implementation follows OpenTelemetry best practices and is ready for deployment to production environments with proper observability infrastructure (OTLP Collector, Tempo, Prometheus, Grafana).
