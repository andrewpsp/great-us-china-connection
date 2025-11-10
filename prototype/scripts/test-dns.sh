#!/bin/bash
set -e

# Test DNS resolution using the Conn3ction CoreDNS deployment

COREDNS_SERVICE=${COREDNS_SERVICE:-coredns}
NAMESPACE=${NAMESPACE:-default}

echo "Testing Conn3ction CoreDNS..."
echo "Using service: $COREDNS_SERVICE in namespace: $NAMESPACE"
echo ""

# Get the CoreDNS service IP
COREDNS_IP=$(kubectl get svc $COREDNS_SERVICE -n $NAMESPACE -o jsonpath='{.spec.clusterIP}')

if [ -z "$COREDNS_IP" ]; then
  echo "Error: Could not find CoreDNS service '$COREDNS_SERVICE' in namespace '$NAMESPACE'"
  exit 1
fi

echo "CoreDNS IP: $COREDNS_IP"
echo ""

# Port-forward to CoreDNS for testing (or use cluster IP from within cluster)
echo "Setting up port-forward to CoreDNS (press Ctrl+C to stop)..."
kubectl port-forward -n $NAMESPACE svc/$COREDNS_SERVICE 5353:53 &
PF_PID=$!

# Wait for port-forward to be ready
sleep 3

# Cleanup function
cleanup() {
  echo ""
  echo "Cleaning up port-forward..."
  kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

echo ""
echo "Testing DNS queries..."
echo ""

# Test queries
echo "1. Query test.example.local (should return 10.1.2.3):"
dig @localhost -p 5353 test.example.local +short || echo "Query failed"
echo ""

echo "2. Query ns1.example.local (should return 10.0.0.1):"
dig @localhost -p 5353 ns1.example.local +short || echo "Query failed"
echo ""

echo "3. Query example.local (should return NS record):"
dig @localhost -p 5353 example.local NS +short || echo "Query failed"
echo ""

echo "4. Query external domain (should forward):"
dig @localhost -p 5353 google.com +short | head -3 || echo "Query failed"
echo ""

echo "Testing complete!"
echo ""
echo "To test from within the cluster, run:"
echo "  kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup test.example.local $COREDNS_IP"
