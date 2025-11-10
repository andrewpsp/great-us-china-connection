# great-us-china-connection

Automated Infrastructure Connection Practices

## Conn3ction DNS Prototype

This repository includes the Conn3ction prototype - a minimal DNS and region-aware system implementing CoreDNS + etcd + a simple DNS API with Helm charts.

### Quick Start

#### Prerequisites

- Go 1.21+ (for building the API)
- Docker (for building images)
- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl
- Helm 3.x
- dig (for testing DNS)

#### Build the API Service

```bash
cd prototype/api
go build -o conn3ction-api .
```

#### Build Docker Image

```bash
cd prototype/api
docker build -t conn3ction-api:latest .
```

For kind clusters:
```bash
./prototype/scripts/build-and-load.sh
```

#### Deploy with Helm

Deploy etcd and CoreDNS:
```bash
helm install coredns-etcd ./prototype/charts/coredns-etcd/
```

Deploy the API service:
```bash
helm install conn3ction-api ./prototype/charts/api/ \
  --set etcd.enabled=true \
  --set etcd.endpoints="coredns-etcd-etcd-0.coredns-etcd-etcd:2379"
```

#### Deploy with kubectl

```bash
kubectl apply -f prototype/k8s/etcd-statefulset.yaml
kubectl apply -f prototype/k8s/coredns-deployment.yaml
kubectl apply -f prototype/k8s/api-deployment.yaml
```

#### Test the API

Port-forward the API service:
```bash
kubectl port-forward svc/conn3ction-api 8080:8080
```

Test health endpoint:
```bash
curl http://localhost:8080/healthz
```

Create a DNS record:
```bash
curl -X POST http://localhost:8080/records \
  -H "Content-Type: application/json" \
  -d '{"name":"app.example.local","type":"A","values":["10.1.2.3"],"ttl":60}'
```

List records:
```bash
curl http://localhost:8080/records
```

#### Test DNS Resolution

Use the test script:
```bash
./prototype/scripts/test-dns.sh
```

Or manually with port-forward:
```bash
kubectl port-forward svc/coredns 5353:53
dig @localhost -p 5353 app.example.local
```

### API Endpoints

- `GET /healthz` - Health check (returns 200 OK)
- `GET /records` - List all DNS records
- `POST /records` - Create a DNS record
- `PUT /records/{name}` - Update a DNS record
- `DELETE /records/{name}` - Delete a DNS record
- `GET /records/{name}` - Get a specific DNS record

### Configuration

#### API Service Environment Variables

- `ETCD_ENDPOINTS` - Comma-separated etcd endpoints (optional, uses in-memory store if not set)
- `ETCD_PREFIX` - Key prefix for records (default: `/conn3ction/records/`)

#### Helm Chart Values

See `prototype/charts/api/values.yaml` and `prototype/charts/coredns-etcd/values.yaml` for configuration options.

### Architecture

See [CONN3CTION_ARCHITECTURE.md](./CONN3CTION_ARCHITECTURE.md) for detailed architecture documentation.

### Development

Run the API locally:
```bash
cd prototype/api
go run .
```

Run with etcd (requires etcd running locally):
```bash
export ETCD_ENDPOINTS="localhost:2379"
go run .
```

Run tests:
```bash
cd prototype/api
go test -v ./...
```

### Next Steps

- [ ] Implement RBAC for API authentication and authorization
- [ ] Add persistence hardening for production deployments
- [ ] Integrate with region-manager for multi-region awareness
- [ ] Add support for additional DNS record types (CNAME, SRV, TXT)
- [ ] Implement geographic routing capabilities
- [ ] Add Prometheus metrics and monitoring
- [ ] Enable DNSSEC support

### Directory Structure

```
prototype/
├── api/                    # Go API service
│   ├── main.go            # Application entry point
│   ├── handlers.go        # HTTP handlers
│   ├── store.go           # Storage implementations
│   ├── go.mod             # Go module definition
│   └── Dockerfile         # Container image
├── charts/                # Helm charts
│   ├── api/              # API service chart
│   └── coredns-etcd/     # CoreDNS + etcd chart
├── k8s/                  # Kubernetes manifests
│   ├── api-deployment.yaml
│   ├── etcd-statefulset.yaml
│   └── coredns-deployment.yaml
└── scripts/              # Helper scripts
    ├── build-and-load.sh
    └── test-dns.sh
```

### Contributing

When contributing to the Conn3ction prototype:
1. Ensure Go code is formatted with `gofmt`
2. Run `go vet` to check for issues
3. Test locally before submitting PRs
4. Update documentation for new features

### License

See repository license file for details.

