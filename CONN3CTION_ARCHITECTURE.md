# Conn3ction Architecture

## Overview

Conn3ction (formerly PolyCloud) is a DNS-based, region-aware routing and service discovery system designed to enable seamless connectivity across geographically distributed infrastructure. The system leverages CoreDNS with etcd as a distributed backend to provide dynamic DNS record management and intelligent traffic routing.

## Core Components

### 1. DNS API Service
A lightweight Go-based REST API that provides CRUD operations for DNS records:
- **Technology**: Go (golang), gorilla/mux for routing
- **Storage Backends**: 
  - In-memory store (default, for development)
  - etcd (for production, distributed state)
- **Endpoints**:
  - `GET /healthz` - Health check
  - `GET /records` - List all DNS records
  - `POST /records` - Create new DNS record
  - `PUT /records/{name}` - Update existing record
  - `DELETE /records/{name}` - Delete record
- **Record Format**: `{"name":"app.example.local","type":"A","values":["10.1.2.3"],"ttl":60}`

### 2. CoreDNS
An extensible DNS server that serves as the primary DNS resolver:
- **Plugin Configuration**: 
  - etcd plugin for dynamic record lookup from etcd backend
  - file plugin for static zone files (development/fallback)
  - forward plugin for upstream DNS resolution
  - cache plugin for performance optimization
- **Zone Management**: Serves configured zones (e.g., `example.local.`) with records stored in etcd under `/conn3ction/zones`

### 3. etcd
A distributed key-value store providing:
- **Persistent Storage**: DNS records stored under `/conn3ction/records/` and `/conn3ction/zones/`
- **Consistency**: Strong consistency guarantees for DNS data
- **Watch API**: Real-time updates for DNS changes
- **High Availability**: Can be deployed as a multi-node cluster (prototype uses single node)

## Data Flow

### Record Creation Flow
```
1. Client → POST /records → DNS API
2. DNS API → Validate record → Store in backend (Memory or etcd)
3. If etcd: DNS API → Put /conn3ction/records/{name} → etcd
4. CoreDNS (etcd plugin) → Watch etcd → Auto-discover new records
5. DNS queries → CoreDNS → etcd lookup → Return response
```

### DNS Resolution Flow
```
1. Client → DNS Query (e.g., app.example.local) → CoreDNS
2. CoreDNS → Check cache
3. If not cached:
   - Query etcd backend at /conn3ction/zones/example.local/app
   - Or lookup from static zone file
4. Return A record (e.g., 10.1.2.3) to client
5. Cache for TTL duration
```

## Deployment Architecture

### Development (Minikube/Kind)
```
┌─────────────────────────────────────┐
│         Kubernetes Cluster          │
├─────────────────────────────────────┤
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │  DNS API     │  │   CoreDNS   │ │
│  │  (Pod)       │  │   (Pod)     │ │
│  └──────┬───────┘  └──────┬──────┘ │
│         │                  │        │
│         │  ┌──────────────┐        │
│         └──│    etcd      │◄───────┘
│            │ (StatefulSet)│        │
│            └──────────────┘        │
│                                     │
└─────────────────────────────────────┘
```

### Production (Multi-Region)
```
┌────────────────────┐         ┌────────────────────┐
│   Region: US-West  │         │   Region: CN-East  │
├────────────────────┤         ├────────────────────┤
│                    │         │                    │
│  ┌───────────────┐ │         │  ┌───────────────┐ │
│  │ DNS API       │ │         │  │ DNS API       │ │
│  └───────┬───────┘ │         │  └───────┬───────┘ │
│          │         │         │          │         │
│  ┌───────▼───────┐ │         │  ┌───────▼───────┐ │
│  │   CoreDNS     │ │         │  │   CoreDNS     │ │
│  └───────┬───────┘ │         │  └───────┬───────┘ │
│          │         │         │          │         │
└──────────┼─────────┘         └──────────┼─────────┘
           │                              │
           │      ┌──────────────┐        │
           └──────►  etcd Cluster◄────────┘
                  │  (Multi-node) │
                  └──────────────┘
```

## Key Design Principles

### 1. Separation of Concerns
- **DNS API**: Record management and business logic
- **CoreDNS**: DNS protocol handling and resolution
- **etcd**: Distributed state and persistence

### 2. Flexibility
- Support for multiple storage backends (in-memory for dev, etcd for prod)
- Configurable via environment variables
- Helm charts for easy deployment customization

### 3. Scalability
- Stateless DNS API (can be horizontally scaled)
- CoreDNS can be scaled across multiple replicas
- etcd provides distributed backend (can be clustered)

### 4. Region Awareness (Future)
- DNS records can include region metadata
- Intelligent routing based on client location
- Latency-based or geo-proximity routing

## etcd Key Structure

### Records (API Storage)
```
/conn3ction/records/{fqdn} → JSON record data
Example:
/conn3ction/records/app.example.local → {"name":"app.example.local","type":"A","values":["10.1.2.3"],"ttl":60}
```

### Zones (CoreDNS Backend)
```
/conn3ction/zones/{zone}/{hostname}/{type} → DNS record value
Example:
/conn3ction/zones/example.local/app/A → 10.1.2.3
```

## Configuration

### DNS API Environment Variables
- `ETCD_ENDPOINTS`: Comma-separated etcd endpoints (e.g., "etcd:2379")
- `ETCD_PREFIX`: Key prefix for records (default: "/conn3ction/records")
- `PORT`: API listen port (default: "8080")

### CoreDNS Corefile
```
example.local. {
    log
    errors
    etcd {
        path /conn3ction/zones
        endpoint etcd:2379
    }
    forward . /etc/resolv.conf
    cache 30
}
```

## Security Considerations (To Be Implemented)

1. **Authentication**: mTLS for etcd client connections
2. **Authorization**: RBAC for API access control
3. **Encryption**: TLS for DNS over TCP/TLS (DoT)
4. **Network Policies**: Kubernetes NetworkPolicies to restrict traffic
5. **Secrets Management**: Use Kubernetes Secrets for credentials

## Monitoring and Observability (Future)

1. **Metrics**: Prometheus metrics from CoreDNS and API
2. **Logging**: Structured logging with correlation IDs
3. **Tracing**: Distributed tracing for DNS resolution path
4. **Alerting**: Alerts for DNS failures, etcd unavailability

## Next Steps

1. **RBAC Implementation**: Add authentication and authorization to API
2. **Persistence Hardening**: Production-grade etcd configuration with persistence
3. **Region Manager Integration**: Service that manages region-based routing
4. **Health Checks**: Advanced health checking and failover logic
5. **Multi-Cluster Support**: Federation across multiple Kubernetes clusters
6. **Performance Optimization**: Caching strategies, connection pooling
7. **Documentation**: API documentation (OpenAPI/Swagger)

## Migration from PolyCloud

All references to PolyCloud have been rebranded to Conn3ction:
- Project name: PolyCloud → Conn3ction
- etcd key prefixes: `/polycloud/*` → `/conn3ction/*`
- Service names: polycloud-api → conn3ction-api
- Documentation and configuration files updated

## References

- [CoreDNS Documentation](https://coredns.io/)
- [etcd Documentation](https://etcd.io/)
- [Kubernetes DNS Specification](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
