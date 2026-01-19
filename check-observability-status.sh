#!/bin/bash

# Check OpenTelemetry Observability Status
# This script helps diagnose observability issues

set -e

echo "🔍 OpenTelemetry Observability Status Check"
echo "==========================================="
echo ""

# Check if logged into OpenShift
if ! oc whoami &> /dev/null; then
    echo "❌ Not logged into OpenShift. Please run: oc login"
    exit 1
fi

echo "✓ Logged into OpenShift as: $(oc whoami)"
echo ""

# Check if ConfigMap has OTEL config
echo "📋 Checking ConfigMap for OpenTelemetry configuration..."
if oc get configmap wailsalutem-backend-config -n wailsalutem-suite -o yaml | grep -q "OTEL_EXPORTER_OTLP_ENDPOINT"; then
    echo "✅ ConfigMap has OTEL configuration"
    echo ""
    echo "Current OTLP Endpoint:"
    oc get configmap wailsalutem-backend-config -n wailsalutem-suite -o yaml | grep "OTEL_EXPORTER_OTLP_ENDPOINT"
else
    echo "❌ ConfigMap is missing OTEL configuration"
    echo "   Run: ./deploy-observability-fix.sh"
    exit 1
fi

echo ""
echo "🔍 Checking if observability namespace exists..."
if oc get namespace observability &> /dev/null; then
    echo "✅ Namespace 'observability' exists"
    echo ""
    echo "Services in observability namespace:"
    oc get services -n observability
else
    echo "⚠️  Namespace 'observability' not found"
    echo "   Your observability stack might be in a different namespace"
    echo ""
    echo "Searching for common observability services..."
    oc get services --all-namespaces | grep -E "(tempo|otel|prometheus|grafana)" || echo "   No observability services found"
fi

echo ""
echo "🚀 Checking organization-service deployment..."
if oc get deployment organization-service -n wailsalutem-suite &> /dev/null; then
    echo "✅ Deployment exists"
    echo ""
    echo "Deployment status:"
    oc get deployment organization-service -n wailsalutem-suite
    echo ""
    echo "Pod status:"
    oc get pods -n wailsalutem-suite -l app=organization-service
else
    echo "❌ Deployment not found"
    exit 1
fi

echo ""
echo "📝 Recent logs (checking for OpenTelemetry initialization)..."
echo "---"
oc logs deployment/organization-service -n wailsalutem-suite --tail=50 | grep -A 5 -B 5 -i "opentelemetry\|otel" || echo "No OpenTelemetry logs found (service might not have restarted yet)"
echo "---"

echo ""
echo "💡 Next Steps:"
echo ""
if oc logs deployment/organization-service -n wailsalutem-suite --tail=50 | grep -q "✓ OpenTelemetry tracer provider initialized"; then
    echo "✅ OpenTelemetry is initialized successfully!"
    echo ""
    echo "🎯 Check Grafana for traces:"
    echo "   1. Go to: https://grafana-observability.apps.inholland-minor.openshift.eu"
    echo "   2. Navigate to Explore → Tempo"
    echo "   3. Search: {service.name=\"organization-service\"}"
    echo ""
    echo "📊 Generate test traffic:"
    echo "   export TOKEN=\$(./get-token.sh)"
    echo "   curl -H \"Authorization: Bearer \$TOKEN\" https://your-service-url/organizations"
elif oc logs deployment/organization-service -n wailsalutem-suite --tail=50 | grep -q "Warning: failed to initialize OpenTelemetry"; then
    echo "⚠️  OpenTelemetry initialization failed"
    echo ""
    echo "Possible causes:"
    echo "   1. OTLP endpoint is incorrect in ConfigMap"
    echo "   2. OTLP Collector is not running"
    echo "   3. Network connectivity issue"
    echo ""
    echo "🔧 Troubleshooting:"
    echo "   1. Check the endpoint in ConfigMap matches your OTLP collector service"
    echo "   2. Verify OTLP collector is running: oc get pods -n observability"
    echo "   3. Check service logs: oc logs -f deployment/organization-service -n wailsalutem-suite"
    echo ""
    echo "📖 See OBSERVABILITY_DEPLOYMENT_FIX.md for detailed troubleshooting"
else
    echo "⚠️  No OpenTelemetry logs found"
    echo ""
    echo "This might mean:"
    echo "   1. Service hasn't restarted since ConfigMap update"
    echo "   2. Service is not running"
    echo ""
    echo "🔧 Try:"
    echo "   ./deploy-observability-fix.sh"
fi

echo ""
