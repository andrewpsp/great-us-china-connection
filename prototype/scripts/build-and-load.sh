#!/bin/bash
set -e

# Build and load Conn3ction prototype images into kind cluster

echo "Building Conn3ction API Docker image..."
cd "$(dirname "$0")/../api"
docker build -t conn3ction-api:latest .

# Check if kind cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^kind$"; then
  echo "No kind cluster named 'kind' found. Please create one first with:"
  echo "  kind create cluster"
  exit 1
fi

echo "Loading image into kind cluster..."
kind load docker-image conn3ction-api:latest

echo "Done! Image conn3ction-api:latest is now available in the kind cluster."
echo ""
echo "To deploy with kubectl:"
echo "  kubectl apply -f ../k8s/"
echo ""
echo "To deploy with Helm:"
echo "  helm install conn3ction-api ../charts/api/"
echo "  helm install conn3ction-dns ../charts/coredns-etcd/"
