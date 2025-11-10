#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==> Testing CoreDNS resolution in cluster"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl."
    exit 1
fi

# Check if dig is available
if ! command -v dig &> /dev/null; then
    echo "Warning: dig not found. Install dnsutils for dig command."
    echo "Attempting to use nslookup instead..."
    DIG_CMD="nslookup"
else
    DIG_CMD="dig"
fi

# Get CoreDNS service IP
COREDNS_SVC=$(kubectl get svc -l app=coredns -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -z "$COREDNS_SVC" ]; then
    echo "Error: CoreDNS service not found. Please deploy CoreDNS first."
    echo "  kubectl apply -f ../k8s/coredns-deployment.yaml"
    exit 1
fi

COREDNS_IP=$(kubectl get svc "$COREDNS_SVC" -o jsonpath='{.spec.clusterIP}')

echo -e "${GREEN}Found CoreDNS service:${NC} $COREDNS_SVC at $COREDNS_IP"
echo ""

# Function to test DNS resolution using a temporary pod
test_dns() {
    local hostname=$1
    echo "Testing DNS query for: $hostname"
    
    kubectl run -it --rm dns-test-pod --image=nicolaka/netshoot --restart=Never -- \
        dig @${COREDNS_IP} ${hostname} +short 2>/dev/null || \
    kubectl run -it --rm dns-test-pod --image=busybox:latest --restart=Never -- \
        nslookup ${hostname} ${COREDNS_IP} 2>/dev/null
    
    echo ""
}

# Test some example records
echo -e "${YELLOW}Testing example.local zone records:${NC}"
test_dns "app.example.local"
test_dns "db.example.local"

echo ""
echo -e "${GREEN}==> DNS testing complete!${NC}"
echo ""
echo "To test manually from inside the cluster, run:"
echo "  kubectl run -it --rm dns-test --image=nicolaka/netshoot --restart=Never -- dig @${COREDNS_IP} app.example.local"
echo ""
echo "To port-forward CoreDNS for local testing:"
echo "  kubectl port-forward svc/${COREDNS_SVC} 5353:53"
echo "  dig @localhost -p 5353 app.example.local"
