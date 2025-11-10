# great-us-china-connection
Automated Infrastructure Connection Practices

---

## PolyCloud Prototype

This repository contains a minimal prototype of the PolyCloud DNS system - a region-aware DNS solution for multi-cloud environments.

### Quick Start

#### Prerequisites

- Go 1.24+ (for building the API)
- Docker (for containerization)
- kubectl (for Kubernetes deployment)
- kind or minikube (for local testing)
- helm (optional, for Helm chart deployment)

#### Building Locally

```bash
# Build the API
cd prototype/api
go build -o polycloud-api

# Run locally (in-memory mode)
./polycloud-api

# Or run with etcd backend
export ETCD_ENDPOINTS=localhost:2379
export ETCD_PREFIX=/polycloud/records
./polycloud-api
```

#### Testing the API

```bash
# Health check
curl http://localhost:8080/healthz

# Create a DNS record
curl -X POST http://localhost:8080/records \
  -H "Content-Type: application/json" \
  -d '{"name":"app.example.local","type":"A","values":["10.1.2.3"],"ttl":60}'

# List all records
curl http://localhost:8080/records

# Get a specific record
curl http://localhost:8080/records/app.example.local

# Update a record
curl -X PUT http://localhost:8080/records/app.example.local \
  -H "Content-Type: application/json" \
  -d '{"type":"A","values":["10.1.2.4"],"ttl":120}'

# Delete a record
curl -X DELETE http://localhost:8080/records/app.example.local
```

#### Deploying to Kubernetes

##### Option 1: Using kubectl

```bash
# Build and load image into kind cluster
cd prototype/scripts
./build-and-load.sh

# Deploy CoreDNS
kubectl apply -f ../k8s/coredns-deployment.yaml

# Deploy the API
kubectl apply -f ../k8s/api-deployment.yaml

# Optional: Deploy etcd
kubectl apply -f ../k8s/etcd-statefulset.yaml
```

##### Option 2: Using Helm

```bash
# Install CoreDNS with ConfigMap backend (development)
helm install coredns-dev prototype/charts/coredns-etcd/ --set useEtcd=false

# Or install CoreDNS with etcd backend (production)
helm install coredns-prod prototype/charts/coredns-etcd/ --set useEtcd=true

# Install the API
helm install polycloud-api prototype/charts/api/
```

#### Testing DNS Resolution

```bash
# Run the test script
cd prototype/scripts
./test-dns.sh

# Or test manually
kubectl run -it --rm dns-test --image=nicolaka/netshoot --restart=Never -- \
  dig @<coredns-service-ip> app.example.local

# Port-forward for local testing
kubectl port-forward svc/coredns 5353:53
dig @localhost -p 5353 app.example.local
```

### Project Structure

```
prototype/
├── api/                      # Go DNS API service
│   ├── main.go              # HTTP server and routing
│   ├── handlers.go          # API handlers
│   ├── store.go             # Storage implementations
│   ├── store_test.go        # Unit tests
│   ├── Dockerfile           # Container image
│   └── go.mod               # Go dependencies
├── charts/                   # Helm charts
│   ├── api/                 # API service chart
│   └── coredns-etcd/        # CoreDNS + etcd chart
├── k8s/                      # Kubernetes manifests
│   ├── api-deployment.yaml
│   ├── coredns-deployment.yaml
│   └── etcd-statefulset.yaml
└── scripts/                  # Helper scripts
    ├── build-and-load.sh    # Build and load into kind
    └── test-dns.sh          # Test DNS resolution
```

### Architecture

See [POLYCLOUD_ARCHITECTURE.md](POLYCLOUD_ARCHITECTURE.md) for detailed architecture documentation.

### Development

```bash
# Run tests
cd prototype/api
go test -v ./...

# Run with race detection
go test -race ./...

# Build Docker image
docker build -t polycloud-api:latest .
```

### Next Steps

- [ ] Add authentication and authorization
- [ ] Implement region-aware routing
- [ ] Add support for additional DNS record types (CNAME, SRV, TXT)
- [ ] Implement health checks and service discovery
- [ ] Add monitoring and observability
- [ ] Improve persistence and high availability

### Contributing

This is a prototype for demonstration and development purposes. For production use, additional hardening, security, and reliability features should be implemented.

---

## Original Infrastructure Automation

The remainder of this repository contains Ansible playbooks and Terraform configurations for automated infrastructure provisioning.
