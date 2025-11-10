#!/bin/bash
# Build and load Conn3ction API image into kind cluster

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="$SCRIPT_DIR/../api"

echo "Building Conn3ction API Docker image..."
cd "$API_DIR"
docker build -t conn3ction-api:latest .

# Check if kind is available and load into kind cluster
if command -v kind &> /dev/null; then
    echo "Loading image into kind cluster..."
    kind load docker-image conn3ction-api:latest || echo "Warning: Failed to load image into kind cluster. Make sure a cluster is running."
else
    echo "kind not found. Skipping kind cluster load."
fi

echo "Build complete! Image: conn3ction-api:latest"
echo ""
echo "To deploy to Kubernetes:"
echo "  kubectl apply -f $SCRIPT_DIR/../k8s/api-deployment.yaml"
echo ""
echo "To deploy with Helm:"
echo "  helm install conn3ction-api $SCRIPT_DIR/../charts/api/"
