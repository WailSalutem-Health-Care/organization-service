package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds OpenTelemetry configuration
type Config struct {
	ServiceName      string
	ServiceNamespace string
	ServiceVersion   string
	Environment      string
	OTLPEndpoint     string
	TracesSampler    string
	MetricsInterval  time.Duration
}

// LoadConfig loads OpenTelemetry configuration from environment variables
func LoadConfig() Config {
	// Get OTLP endpoint with default
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4317"
	}

	// Get service name with default
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "organization-service"
	}

	// Get service namespace with default
	serviceNamespace := os.Getenv("OTEL_SERVICE_NAMESPACE")
	if serviceNamespace == "" {
		serviceNamespace = "wailsalutem"
	}

	// Get service version with default
	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	// Get environment with default
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "production"
	}

	// Get traces sampler with default
	tracesSampler := os.Getenv("OTEL_TRACES_SAMPLER")
	if tracesSampler == "" {
		tracesSampler = "always_on"
	}

	// Get metrics export interval with default
	metricsIntervalStr := os.Getenv("OTEL_METRICS_EXPORT_INTERVAL")
	metricsInterval := 30 * time.Second
	if metricsIntervalStr != "" {
		if duration, err := time.ParseDuration(metricsIntervalStr); err == nil {
			metricsInterval = duration
		}
	}

	return Config{
		ServiceName:      serviceName,
		ServiceNamespace: serviceNamespace,
		ServiceVersion:   serviceVersion,
		Environment:      environment,
		OTLPEndpoint:     otlpEndpoint,
		TracesSampler:    tracesSampler,
		MetricsInterval:  metricsInterval,
	}
}

// Provider holds the OpenTelemetry providers
type Provider struct {
	TracerProvider *trace.TracerProvider
	MeterProvider  *metric.MeterProvider
	config         Config
}

// InitProvider initializes OpenTelemetry tracer and meter providers
// It fails gracefully if the OTLP collector is unavailable
func InitProvider(ctx context.Context, cfg Config) (*Provider, error) {
	log.Printf("Initializing OpenTelemetry with endpoint: %s", cfg.OTLPEndpoint)

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.ServiceNamespace),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize trace provider
	tracerProvider, err := initTracerProvider(ctx, cfg, res)
	if err != nil {
		log.Printf("Warning: failed to initialize tracer provider: %v", err)
		log.Println("Service will continue without distributed tracing")
		tracerProvider = nil
	} else {
		// Set global tracer provider
		otel.SetTracerProvider(tracerProvider)
		log.Println("✓ OpenTelemetry tracer provider initialized")
	}

	// Initialize meter provider
	meterProvider, err := initMeterProvider(ctx, cfg, res)
	if err != nil {
		log.Printf("Warning: failed to initialize meter provider: %v", err)
		log.Println("Service will continue without metrics export")
		meterProvider = nil
	} else {
		// Set global meter provider
		otel.SetMeterProvider(meterProvider)
		log.Println("✓ OpenTelemetry meter provider initialized")
	}

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		config:         cfg,
	}, nil
}

// initTracerProvider initializes the trace provider with OTLP exporter
func initTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*trace.TracerProvider, error) {
	// Create OTLP trace exporter with timeout and retry
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Determine sampler based on configuration
	var sampler trace.Sampler
	switch cfg.TracesSampler {
	case "always_on":
		sampler = trace.AlwaysSample()
	case "always_off":
		sampler = trace.NeverSample()
	case "traceidratio":
		sampler = trace.TraceIDRatioBased(0.1) // Sample 10% of traces
	default:
		sampler = trace.AlwaysSample()
	}

	// Create tracer provider with batch span processor
	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(sampler),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithMaxExportBatchSize(512),
		),
	)

	return tracerProvider, nil
}

// initMeterProvider initializes the meter provider with OTLP exporter
func initMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*metric.MeterProvider, error) {
	// Create OTLP metric exporter with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlpmetricgrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create meter provider with periodic reader
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(cfg.MetricsInterval),
		)),
	)

	return meterProvider, nil
}

// Shutdown gracefully shuts down the OpenTelemetry providers
func (p *Provider) Shutdown(ctx context.Context) error {
	log.Println("Shutting down OpenTelemetry providers...")

	var err error

	// Shutdown tracer provider
	if p.TracerProvider != nil {
		if shutdownErr := p.TracerProvider.Shutdown(ctx); shutdownErr != nil {
			log.Printf("Error shutting down tracer provider: %v", shutdownErr)
			err = shutdownErr
		} else {
			log.Println("✓ Tracer provider shut down successfully")
		}
	}

	// Shutdown meter provider
	if p.MeterProvider != nil {
		if shutdownErr := p.MeterProvider.Shutdown(ctx); shutdownErr != nil {
			log.Printf("Error shutting down meter provider: %v", shutdownErr)
			if err == nil {
				err = shutdownErr
			}
		} else {
			log.Println("✓ Meter provider shut down successfully")
		}
	}

	return err
}
