# Conn3ction - Great US-China Connection
Automated Infrastructure Connection Practices

## Overview

This repository contains infrastructure automation tools and the **Conn3ction prototype** - a DNS-based, region-aware routing and service discovery system designed for distributed infrastructure.

## Conn3ction Prototype

The `prototype/` directory contains a minimal implementation of Conn3ction featuring:
- **DNS API**: Go-based REST API for managing DNS records (A records)
- **CoreDNS**: DNS server with etcd backend for dynamic record resolution
- **etcd**: Distributed key-value store for DNS data persistence
- **Helm Charts**: Deployment charts for Kubernetes
- **Example Manifests**: Direct kubectl deployment examples

### Quick Start (Local Development)

#### Prerequisites
- Go 1.21+ (for building the API)
- Docker (for containerization)
- Kubernetes cluster (minikube, kind, or similar)
- kubectl and helm CLI tools
- dig (for DNS testing)

#### Build the API

```bash
cd prototype/api
go mod download
go build -v .
```

#### Run locally (in-memory mode)

```bash
cd prototype/api
./api
# API will start on http://localhost:8080
```

#### Test the API

```bash
# Health check
curl http://localhost:8080/healthz

# Create a DNS record
curl -X POST http://localhost:8080/records \
  -H "Content-Type: application/json" \
  -d '{"name":"app.example.local","type":"A","values":["10.1.2.3"],"ttl":60}'

# List all records
curl http://localhost:8080/records

# Get specific record
curl http://localhost:8080/records/app.example.local

# Update record
curl -X PUT http://localhost:8080/records/app.example.local \
  -H "Content-Type: application/json" \
  -d '{"name":"app.example.local","type":"A","values":["10.1.2.4"],"ttl":120}'

# Delete record
curl -X DELETE http://localhost:8080/records/app.example.local
```

### Kubernetes Deployment

#### Option 1: Using Helm Charts

```bash
# Deploy CoreDNS with etcd
helm install conn3ction-dns prototype/charts/coredns-etcd/

# Deploy the API (with etcd backend)
helm install conn3ction-api prototype/charts/api/ \
  --set env.ETCD_ENDPOINTS="etcd:2379" \
  --set env.ETCD_PREFIX="/conn3ction/records"

# Or deploy API with in-memory storage (for testing)
helm install conn3ction-api prototype/charts/api/
```

#### Option 2: Using kubectl

```bash
# Deploy all components
kubectl apply -f prototype/k8s/

# Or deploy individually
kubectl apply -f prototype/k8s/etcd-statefulset.yaml
kubectl apply -f prototype/k8s/coredns-deployment.yaml
kubectl apply -f prototype/k8s/api-deployment.yaml
```

#### Option 3: Using kind (local development)

```bash
# Create a kind cluster
kind create cluster

# Build and load the API image
cd prototype/scripts
./build-and-load.sh

# Deploy with kubectl
kubectl apply -f ../k8s/
```

### Testing DNS Resolution

```bash
# Using the test script
cd prototype/scripts
./test-dns.sh

# Manual testing with dig (requires port-forward)
kubectl port-forward svc/coredns 5353:53
dig @localhost -p 5353 test.example.local

# Testing from within the cluster
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nslookup test.example.local coredns
```

### Configuration

#### DNS API Environment Variables
- `ETCD_ENDPOINTS`: Comma-separated etcd endpoints (default: in-memory mode)
- `ETCD_PREFIX`: etcd key prefix (default: `/conn3ction/records`)
- `PORT`: API listen port (default: `8080`)

#### Helm Chart Values
See `prototype/charts/api/values.yaml` and `prototype/charts/coredns-etcd/values.yaml` for all available configuration options.

### Architecture

For detailed architecture documentation, see [CONN3CTION_ARCHITECTURE.md](./CONN3CTION_ARCHITECTURE.md).

Key components:
- **DNS API**: Manages DNS records via REST API, stores in etcd or memory
- **CoreDNS**: Resolves DNS queries using etcd backend or static zone files
- **etcd**: Provides distributed, consistent storage for DNS records

### Project Structure

```
prototype/
├── api/                    # Go DNS API service
│   ├── main.go            # Entry point and HTTP server
│   ├── handlers.go        # HTTP request handlers
│   ├── store.go           # Storage interface and implementations
│   ├── Dockerfile         # Container image
│   └── go.mod             # Go module dependencies
├── charts/                # Helm charts
│   ├── api/              # DNS API chart
│   └── coredns-etcd/     # CoreDNS + etcd chart
├── k8s/                   # Example Kubernetes manifests
│   ├── api-deployment.yaml
│   ├── etcd-statefulset.yaml
│   └── coredns-deployment.yaml
└── scripts/               # Helper scripts
    ├── build-and-load.sh # Build and load into kind
    └── test-dns.sh        # DNS testing script
```

### Development Notes

- The prototype uses **in-memory storage by default** for easy local development
- Set `ETCD_ENDPOINTS` environment variable to enable etcd backend
- CoreDNS can be configured with or without etcd (see `useEtcd` in values.yaml)
- Static zone files are supported for development without etcd

### Next Steps

- [ ] Add authentication and RBAC to the API
- [ ] Implement persistence hardening for production
- [ ] Region-manager service for intelligent routing
- [ ] Multi-region DNS synchronization
- [ ] Monitoring and metrics (Prometheus)
- [ ] API documentation (OpenAPI/Swagger)

### Troubleshooting

**API won't start:**
- Check if port 8080 is already in use
- Verify etcd connectivity if `ETCD_ENDPOINTS` is set

**DNS queries fail:**
- Ensure CoreDNS pod is running: `kubectl get pods -l app=coredns`
- Check CoreDNS logs: `kubectl logs -l app=coredns`
- Verify zone configuration in ConfigMap

**Helm install fails:**
- Validate chart: `helm template prototype/charts/api/`
- Check for resource conflicts: `kubectl get all`

## License

See LICENSE file for details.
