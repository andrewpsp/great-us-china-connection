#!/bin/bash
# Test DNS resolution using the CoreDNS service

set -e

NAMESPACE="${NAMESPACE:-default}"
COREDNS_SERVICE="${COREDNS_SERVICE:-coredns}"
ZONE="${ZONE:-example.local}"

echo "Testing Conn3ction DNS resolution..."
echo "Namespace: $NAMESPACE"
echo "Service: $COREDNS_SERVICE"
echo "Zone: $ZONE"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl."
    exit 1
fi

# Get the CoreDNS service ClusterIP
DNS_IP=$(kubectl get svc "$COREDNS_SERVICE" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null)

if [ -z "$DNS_IP" ]; then
    echo "Error: Could not find service $COREDNS_SERVICE in namespace $NAMESPACE"
    exit 1
fi

echo "CoreDNS service IP: $DNS_IP"
echo ""

# Test DNS resolution from within the cluster
echo "Running DNS query test from a temporary pod..."
kubectl run -it --rm dns-test --image=busybox:latest --restart=Never -- \
    nslookup "app.$ZONE" "$DNS_IP" || echo "Test pod completed"

echo ""
echo "Alternative test using port-forward:"
echo "  kubectl port-forward svc/$COREDNS_SERVICE 5353:53"
echo "  dig @localhost -p 5353 app.$ZONE"
echo ""
echo "To test the API:"
echo "  kubectl port-forward svc/conn3ction-api 8080:8080"
echo "  curl http://localhost:8080/healthz"
echo "  curl http://localhost:8080/records"
