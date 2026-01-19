#!/bin/bash

# Deploy OpenTelemetry Configuration Fix
# This script updates the ConfigMap and restarts the service

set -e

echo "🔧 Deploying OpenTelemetry Configuration Fix"
echo ""

# Check if logged into OpenShift
if ! oc whoami &> /dev/null; then
    echo "❌ Not logged into OpenShift. Please run: oc login"
    exit 1
fi

echo "✓ Logged into OpenShift as: $(oc whoami)"
echo ""

# Apply the updated ConfigMap
echo "📝 Applying updated ConfigMap with OpenTelemetry configuration..."
oc apply -f k8s/config/configmap.yml

echo ""
echo "🔄 Restarting organization-service pods to pick up new configuration..."
oc rollout restart deployment/organization-service -n wailsalutem-suite

echo ""
echo "⏳ Waiting for rollout to complete..."
oc rollout status deployment/organization-service -n wailsalutem-suite

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Check the logs to verify OpenTelemetry connection:"
echo "   oc logs -f deployment/organization-service -n wailsalutem-suite"
echo ""
echo "You should see:"
echo "   ✓ OpenTelemetry tracer provider initialized"
echo "   ✓ OpenTelemetry meter provider initialized"
echo ""
echo "🎯 Now check Grafana for traces!"
echo "   https://grafana-observability.apps.inholland-minor.openshift.eu"
echo ""
