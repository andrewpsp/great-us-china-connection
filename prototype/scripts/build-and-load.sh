#!/bin/bash
set -e

echo "==> Building PolyCloud API Docker image..."
cd "$(dirname "$0")/../api"
docker build -t polycloud-api:latest .

echo "==> Checking for kind cluster..."
if command -v kind &> /dev/null; then
    CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        echo "==> Loading image into kind cluster: ${CLUSTER_NAME}"
        kind load docker-image polycloud-api:latest --name "${CLUSTER_NAME}"
        echo "==> Image loaded successfully into kind cluster"
    else
        echo "==> No kind cluster named '${CLUSTER_NAME}' found"
        echo "    To load into kind later, run: kind load docker-image polycloud-api:latest --name ${CLUSTER_NAME}"
    fi
else
    echo "==> kind not found, skipping cluster image load"
    echo "    Install kind from: https://kind.sigs.k8s.io/docs/user/quick-start/"
fi

echo ""
echo "==> Build complete! Image: polycloud-api:latest"
echo "    To deploy via kubectl: kubectl apply -f ../k8s/api-deployment.yaml"
echo "    To deploy via helm: helm install polycloud-api ../charts/api/"
