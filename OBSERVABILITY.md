# OpenTelemetry Observability Guide

This document provides comprehensive guidance on the OpenTelemetry instrumentation implemented in the organization-service.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Instrumentation Details](#instrumentation-details)
- [Local Testing Setup](#local-testing-setup)
- [Metrics Reference](#metrics-reference)
- [Traces Reference](#traces-reference)
- [Querying and Analysis](#querying-and-analysis)
- [Troubleshooting](#troubleshooting)

---

## Overview

The organization-service is fully instrumented with OpenTelemetry to provide:

- **Distributed Tracing**: End-to-end request tracing across all components
- **Metrics**: Business and technical metrics for monitoring
- **Context Propagation**: Trace context flows through HTTP, database, and message queues

### Key Features

✅ **HTTP Server Instrumentation** - Automatic tracing of all HTTP requests  
✅ **PostgreSQL Instrumentation** - Query-level tracing and connection pool metrics  
✅ **RabbitMQ Instrumentation** - Message publishing with trace context propagation  
✅ **Keycloak HTTP Client** - External API call tracing  
✅ **Authentication & Authorization** - Auth middleware and permission check tracing  
✅ **Business Metrics** - Organization, patient, and user operation counters  
✅ **Graceful Degradation** - Service continues if OTLP collector is unavailable

---

## Architecture

### Telemetry Stack

```
┌─────────────────────────────────────────────────────────────┐
│                   organization-service                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  OpenTelemetry SDK                                     │ │
│  │  - Tracer Provider (OTLP gRPC)                        │ │
│  │  - Meter Provider (OTLP gRPC)                         │ │
│  │  - Propagators (TraceContext, Baggage)               │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ OTLP/gRPC (port 4317)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              OpenTelemetry Collector                         │
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐ │
│  │   Receivers    │→ │   Processors   │→ │   Exporters   │ │
│  │   (OTLP)       │  │   (Batch)      │  │  (Multiple)   │ │
│  └────────────────┘  └────────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │  Tempo   │    │Prometheus│    │  Loki    │
    │ (Traces) │    │(Metrics) │    │  (Logs)  │
    └──────────┘    └──────────┘    └──────────┘
           │               │               │
           └───────────────┴───────────────┘
                           │
                           ▼
                    ┌──────────┐
                    │ Grafana  │
                    │(Dashboards)
                    └──────────┘
```

### Instrumentation Layers

1. **HTTP Layer** (`internal/http/router.go`)
   - Gorilla Mux middleware (otelmux)
   - Request/response tracing
   - Route-level metrics

2. **Authentication Layer** (`internal/auth/middleware.go`)
   - Token validation tracing
   - Permission check spans
   - Auth failure metrics

3. **Service Layer** (`internal/*/service_instrumented.go`)
   - Business operation spans
   - Operation counters
   - Error tracking

4. **Repository Layer** (automatic via otelsql)
   - SQL query tracing
   - Connection pool metrics
   - Query duration histograms

5. **External Services**
   - Keycloak HTTP client (`internal/auth/keycloak_admin.go`)
   - RabbitMQ publisher (`internal/messaging/rabbitmq.go`)

---

## Configuration

### Environment Variables

All OpenTelemetry configuration is done via environment variables:

```bash
# OTLP Exporter Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317    # OTLP collector endpoint
OTEL_SERVICE_NAME=organization-service         # Service name in traces/metrics
OTEL_SERVICE_NAMESPACE=wailsalutem            # Service namespace
OTEL_SERVICE_VERSION=1.0.0                    # Service version
ENVIRONMENT=production                         # Deployment environment

# Sampling Configuration
OTEL_TRACES_SAMPLER=always_on                 # Options: always_on, always_off, traceidratio
OTEL_METRICS_EXPORT_INTERVAL=30s              # Metrics export interval
```

### Sampling Strategies

- **`always_on`** (default): Sample 100% of traces - use in development/staging
- **`always_off`**: Disable tracing - use for testing without observability
- **`traceidratio`**: Sample 10% of traces - use in high-traffic production

### Resource Attributes

Every trace and metric includes these resource attributes:

```yaml
service.name: organization-service
service.namespace: wailsalutem
service.version: 1.0.0
deployment.environment: production
```

---

## Instrumentation Details

### 1. HTTP Server Instrumentation

**Location**: `cmd/api/main.go`

**Implementation**:
```go
router.Use(otelmux.Middleware("organization-service"))
```

**Captured Data**:
- HTTP method, route, status code
- Request duration
- Error details
- Trace context in response headers

**Example Span**:
```json
{
  "name": "POST /organizations",
  "kind": "server",
  "attributes": {
    "http.method": "POST",
    "http.route": "/organizations",
    "http.status_code": 201,
    "http.target": "/organizations"
  }
}
```

### 2. PostgreSQL Instrumentation

**Location**: `internal/db/postgres.go`

**Implementation**:
```go
db, err := otelsql.Open("postgres", connStr,
    otelsql.WithAttributes(
        semconv.DBSystemPostgreSQL,
        semconv.DBName(dbname),
    ),
)
otelsql.RegisterDBStatsMetrics(db, ...)
```

**Captured Data**:
- SQL queries (sanitized)
- Query duration
- Connection pool stats
- Row counts

**Metrics**:
- `db.client.connections.usage` - Active connections
- `db.client.connections.max` - Max connections
- `db.client.connections.wait_time` - Connection wait time

### 3. RabbitMQ Instrumentation

**Location**: `internal/messaging/rabbitmq.go`

**Implementation**:
```go
ctx, span := tracer.Start(ctx, "rabbitmq.publish",
    trace.WithSpanKind(trace.SpanKindProducer),
)
// Inject trace context into message headers
propagator.Inject(ctx, &rabbitMQCarrier{headers: headers})
```

**Captured Data**:
- Exchange name, routing key
- Message size
- Publish duration
- Trace context propagation

**Example Span**:
```json
{
  "name": "rabbitmq.publish",
  "kind": "producer",
  "attributes": {
    "messaging.system": "rabbitmq",
    "messaging.destination": "wailsalutem.events",
    "messaging.routing_key": "patient.created",
    "messaging.message_payload_size_bytes": 256
  }
}
```

### 4. Keycloak HTTP Client

**Location**: `internal/auth/keycloak_admin.go`

**Implementation**:
```go
httpClient := &http.Client{
    Transport: newOtelTransport(http.DefaultTransport),
}
```

**Captured Data**:
- HTTP method, URL, status
- Request/response duration
- Error details

### 5. Authentication & Authorization

**Location**: `internal/auth/middleware.go`

**Spans**:
- `auth.Middleware` - Token validation
- `auth.RequirePermission` - Permission checks

**Metrics**:
- `auth_failures_total` - Counter with reason label
- `permission_check_duration_ms` - Histogram

**Captured Data**:
```json
{
  "user.id": "uuid",
  "user.email": "user@example.com",
  "user.roles": ["ORG_ADMIN"],
  "organization.id": "org-uuid",
  "permission.required": "patient:create",
  "permission.allowed": true
}
```

### 6. Business Operations

**Locations**: 
- `internal/organization/service_instrumented.go`
- `internal/patient/service_instrumented.go`
- `internal/users/service_instrumented.go`

**Spans**: Each CRUD operation creates a span:
- `organization.CreateOrganization`
- `patient.UpdatePatient`
- `users.DeleteUser`
- etc.

**Metrics**:
- `organization_total{operation="create"}`
- `patient_total{operation="update"}`
- `user_total{operation="delete"}`

---

## Local Testing Setup

### Option 1: Docker Compose with Observability Stack

Create a `docker-compose.observability.yml`:

```yaml
version: '3.8'

services:
  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    container_name: otel-collector
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC receiver
      - "4318:4318"   # OTLP HTTP receiver
      - "8888:8888"   # Prometheus metrics exposed by the collector
      - "8889:8889"   # Prometheus exporter metrics
      - "13133:13133" # health_check extension
    networks:
      - observability

  # Tempo (Traces)
  tempo:
    image: grafana/tempo:latest
    container_name: tempo
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./tempo-config.yaml:/etc/tempo.yaml
      - tempo-data:/tmp/tempo
    ports:
      - "3200:3200"   # Tempo HTTP
      - "4317"        # OTLP gRPC receiver
    networks:
      - observability

  # Prometheus (Metrics)
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"
    networks:
      - observability

  # Grafana (Visualization)
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
      - GF_AUTH_DISABLE_LOGIN_FORM=true
    volumes:
      - ./grafana-datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml
      - grafana-data:/var/lib/grafana
    ports:
      - "3000:3000"
    networks:
      - observability

volumes:
  tempo-data:
  prometheus-data:
  grafana-data:

networks:
  observability:
    driver: bridge
```

### OTEL Collector Configuration

Create `otel-collector-config.yaml`:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024
  
  memory_limiter:
    check_interval: 1s
    limit_mib: 512

exporters:
  # Export traces to Tempo
  otlp/tempo:
    endpoint: tempo:4317
    tls:
      insecure: true

  # Export metrics to Prometheus
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: organization_service

  # Logging exporter for debugging
  logging:
    loglevel: info

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [otlp/tempo, logging]
    
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheus, logging]
```

### Tempo Configuration

Create `tempo-config.yaml`:

```yaml
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/traces

query_frontend:
  search:
    enabled: true
```

### Prometheus Configuration

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'otel-collector'
    static_configs:
      - targets: ['otel-collector:8889']
```

### Grafana Data Sources

Create `grafana-datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true

  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    jsonData:
      tracesToLogs:
        datasourceUid: 'loki'
```

### Starting the Stack

```bash
# Start observability stack
docker-compose -f docker-compose.observability.yml up -d

# Start organization-service
docker-compose up -d

# Or run locally
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
go run cmd/api/main.go
```

### Accessing the UI

- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200

---

## Metrics Reference

### HTTP Metrics

| Metric Name | Type | Labels | Description |
|------------|------|--------|-------------|
| `http_server_requests_total` | Counter | `http_method`, `http_route`, `http_status_code` | Total HTTP requests |
| `http_server_duration_milliseconds` | Histogram | `http_method`, `http_route`, `http_status_code` | HTTP request duration |

### Business Metrics

| Metric Name | Type | Labels | Description |
|------------|------|--------|-------------|
| `organization_total` | Counter | `operation` | Organization operations (create, update, delete, list) |
| `patient_total` | Counter | `operation` | Patient operations (create, update, delete, list) |
| `user_total` | Counter | `operation` | User operations (create, update, delete, list) |

### Authentication Metrics

| Metric Name | Type | Labels | Description |
|------------|------|--------|-------------|
| `auth_failures_total` | Counter | `reason` | Authentication failures (missing_authorization, invalid_token, etc.) |
| `permission_check_duration_ms` | Histogram | `permission`, `allowed` | Permission check duration |

### Database Metrics

| Metric Name | Type | Labels | Description |
|------------|------|--------|-------------|
| `db.client.connections.usage` | Gauge | `db.system`, `db.name`, `state` | Active/idle connections |
| `db.client.connections.max` | Gauge | `db.system`, `db.name` | Maximum connections |
| `db.client.connections.wait_time` | Histogram | `db.system`, `db.name` | Connection wait time |

---

## Traces Reference

### Trace Structure

A typical request trace includes:

```
POST /organization/patients
├── auth.Middleware (token validation)
├── auth.RequirePermission (permission check)
├── patient.CreatePatient (business logic)
│   ├── db:INSERT (PostgreSQL)
│   └── rabbitmq.publish (event publishing)
│       └── keycloak.http_request (create user in Keycloak)
└── HTTP response
```

### Span Attributes

#### HTTP Spans
- `http.method`: GET, POST, PUT, DELETE
- `http.route`: /organizations/{id}
- `http.status_code`: 200, 201, 400, 500
- `http.target`: /organizations/123

#### Database Spans
- `db.system`: postgresql
- `db.name`: organization_db
- `db.statement`: SELECT * FROM ...
- `db.operation`: SELECT, INSERT, UPDATE, DELETE

#### Messaging Spans
- `messaging.system`: rabbitmq
- `messaging.destination`: wailsalutem.events
- `messaging.routing_key`: patient.created
- `messaging.message_payload_size_bytes`: 256

#### User Context
- `user.id`: UUID
- `user.email`: user@example.com
- `user.roles`: [ORG_ADMIN]
- `organization.id`: org-uuid

---

## Querying and Analysis

### Prometheus Queries

#### Request Rate
```promql
# Requests per second by route
rate(http_server_requests_total[5m])

# Error rate (4xx and 5xx)
rate(http_server_requests_total{http_status_code=~"4..|5.."}[5m])
```

#### Latency
```promql
# 95th percentile latency by route
histogram_quantile(0.95, 
  rate(http_server_duration_milliseconds_bucket[5m])
)

# Average latency
rate(http_server_duration_milliseconds_sum[5m]) / 
rate(http_server_duration_milliseconds_count[5m])
```

#### Business Metrics
```promql
# Organizations created per minute
rate(organization_total{operation="create"}[1m]) * 60

# Patient operations breakdown
sum by (operation) (rate(patient_total[5m]))

# Auth failure rate
rate(auth_failures_total[5m])
```

#### Database Metrics
```promql
# Active database connections
db_client_connections_usage{state="used"}

# Connection pool utilization
db_client_connections_usage{state="used"} / 
db_client_connections_max

# Slow queries (>100ms)
histogram_quantile(0.95, 
  rate(db_client_operation_duration_bucket[5m])
) > 0.1
```

### Tempo Queries

#### Find Traces by Service
```
{service.name="organization-service"}
```

#### Find Slow Traces
```
{service.name="organization-service" && duration > 1s}
```

#### Find Errors
```
{service.name="organization-service" && status=error}
```

#### Find by User
```
{service.name="organization-service" && user.id="uuid"}
```

#### Find by Operation
```
{service.name="organization-service" && span.name="patient.CreatePatient"}
```

### Grafana Dashboard Examples

#### Service Overview Dashboard

**Panels**:
1. Request Rate (time series)
2. Error Rate (time series)
3. P50/P95/P99 Latency (time series)
4. Top 10 Slowest Endpoints (table)
5. Error Breakdown by Route (pie chart)

#### Business Metrics Dashboard

**Panels**:
1. Organizations Created (counter)
2. Patients Created (counter)
3. Users Created (counter)
4. Operations by Type (bar chart)
5. Auth Failures (time series)

#### Database Performance Dashboard

**Panels**:
1. Connection Pool Usage (gauge)
2. Query Duration P95 (time series)
3. Slow Queries (table)
4. Queries per Second (time series)

---

## Troubleshooting

### Service Won't Start

**Symptom**: Service fails to start with OTLP connection error

**Solution**: The service is designed to fail gracefully. Check logs:
```
Warning: failed to initialize OpenTelemetry: ...
Service will continue without observability
```

This is expected if the OTLP collector is not running. The service will continue normally.

### No Traces Appearing

**Check**:
1. OTLP collector is running: `curl http://localhost:13133` (health check)
2. Environment variable is set: `echo $OTEL_EXPORTER_OTLP_ENDPOINT`
3. Sampling is enabled: `OTEL_TRACES_SAMPLER=always_on`
4. Check collector logs: `docker logs otel-collector`

### No Metrics Appearing

**Check**:
1. Prometheus is scraping collector: http://localhost:9090/targets
2. Metrics export interval: `OTEL_METRICS_EXPORT_INTERVAL=30s`
3. Check collector metrics endpoint: `curl http://localhost:8889/metrics`

### High Memory Usage

**Solution**: Adjust batch processor settings in collector config:
```yaml
processors:
  batch:
    timeout: 5s
    send_batch_size: 512  # Reduce from 1024
```

### Trace Context Not Propagating

**Check**:
1. Ensure context.Context is passed through all layers
2. Verify propagators are set: `otel.SetTextMapPropagator(...)`
3. Check RabbitMQ message headers include trace context

### Database Spans Missing Query Details

**Note**: Query sanitization is enabled by default for security. To see full queries (dev only):
```go
otelsql.WithSpanOptions(otelsql.SpanOptions{
    DisableQuery: false,  // Show full queries
})
```

---

## Best Practices

### 1. Context Propagation
Always pass `context.Context` through your call chain:
```go
func (s *Service) DoSomething(ctx context.Context) error {
    // ctx carries trace context
    return s.repo.DoSomething(ctx)
}
```

### 2. Span Attributes
Add meaningful attributes to spans:
```go
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.String("operation", "create"),
)
```

### 3. Error Recording
Always record errors in spans:
```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "operation failed")
    return err
}
```

### 4. Metrics Labels
Keep cardinality low on metric labels:
```go
// Good: Limited set of operations
metrics.RecordOperation(ctx, "create")

// Bad: High cardinality (user IDs, timestamps, etc.)
metrics.RecordOperation(ctx, userID)
```

### 5. Sampling in Production
Use sampling for high-traffic services:
```bash
OTEL_TRACES_SAMPLER=traceidratio  # Samples 10% of traces
```

---

## Additional Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [OTLP Specification](https://opentelemetry.io/docs/reference/specification/protocol/)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Prometheus Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)

---

## Support

For issues or questions about observability:
1. Check service logs for warnings
2. Verify OTLP collector is running and healthy
3. Review this documentation
4. Contact the platform team
