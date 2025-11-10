# Conn3ction Architecture

## Overview

Conn3ction (formerly PolyCloud) is a DNS and region-aware system for intelligent service discovery and routing across distributed infrastructure. This document describes the architecture and design principles of the Conn3ction prototype.

## Components

### 1. DNS API Service

The DNS API service provides a RESTful interface for managing DNS records. It supports both in-memory storage (for development) and etcd-backed storage (for production).

**Key Features:**
- RESTful API for CRUD operations on DNS records
- Pluggable storage backend (in-memory or etcd)
- Health check endpoints
- Graceful shutdown
- Cloud-native deployment with Kubernetes

**API Endpoints:**
- `GET /healthz` - Health check
- `GET /records` - List all DNS records
- `POST /records` - Create a new DNS record
- `PUT /records/{name}` - Update an existing DNS record
- `DELETE /records/{name}` - Delete a DNS record
- `GET /records/{name}` - Get a specific DNS record

### 2. CoreDNS

CoreDNS serves as the DNS server, providing resolution for configured zones. It can be configured with:
- etcd backend for dynamic DNS records
- File-based backend for static configurations
- Forwarding to upstream DNS servers for external queries

### 3. etcd

etcd provides distributed storage for DNS records and configuration data. It serves as the backend for both the API service and CoreDNS.

**Key Prefixes:**
- `/conn3ction/records/` - DNS records managed by the API
- `/conn3ction/zones/` - DNS zones for CoreDNS

## Architecture Diagram

```
┌─────────────┐     ┌──────────────┐     ┌─────────┐
│   Clients   │────▶│  DNS API     │────▶│  etcd   │
└─────────────┘     │  Service     │     └─────────┘
                    └──────────────┘          │
                                              │
┌─────────────┐     ┌──────────────┐         │
│ DNS Clients │────▶│  CoreDNS     │◀────────┘
└─────────────┘     └──────────────┘
```

## Deployment Options

### 1. Kubernetes with Helm

The recommended deployment method uses Helm charts:

- `charts/api/` - Deploys the DNS API service
- `charts/coredns-etcd/` - Deploys CoreDNS and optionally etcd

### 2. Direct Kubernetes Manifests

For simpler deployments, raw Kubernetes manifests are available in `k8s/`:

- `api-deployment.yaml` - API service
- `etcd-statefulset.yaml` - etcd cluster
- `coredns-deployment.yaml` - CoreDNS server

## Storage Backends

### In-Memory Store

For development and testing, the API service can use an in-memory store:
- No external dependencies
- Fast and simple
- Data is lost on restart

### etcd Store

For production deployments, etcd provides:
- Persistent storage
- Distributed consensus
- Watch capabilities for real-time updates
- High availability

## Configuration

### Environment Variables

**API Service:**
- `ETCD_ENDPOINTS` - Comma-separated list of etcd endpoints (optional)
- `ETCD_PREFIX` - Key prefix for DNS records (default: `/conn3ction/records/`)

**CoreDNS:**
Configured via Corefile (see Helm charts and k8s manifests)

## Future Enhancements

1. **Region-Aware Routing**
   - Geographic routing based on client location
   - Multi-region DNS resolution
   - Health-based failover

2. **RBAC and Authentication**
   - API authentication and authorization
   - Role-based access control
   - Audit logging

3. **Advanced DNS Features**
   - Support for additional record types (CNAME, SRV, TXT, etc.)
   - DNSSEC support
   - DNS-based load balancing

4. **Monitoring and Observability**
   - Prometheus metrics
   - Distributed tracing
   - Structured logging

5. **High Availability**
   - Multi-replica etcd clusters
   - CoreDNS clustering
   - API service horizontal scaling

## Record Format

DNS records are represented in JSON format:

```json
{
  "name": "app.example.local",
  "type": "A",
  "values": ["10.1.2.3"],
  "ttl": 60
}
```

## Security Considerations

1. **Network Policies**: Restrict access to etcd and API services
2. **TLS**: Enable TLS for etcd client connections
3. **Authentication**: Implement API key or OAuth2 authentication
4. **RBAC**: Define roles for read-only vs. read-write access
5. **Rate Limiting**: Protect API from abuse

## Development and Testing

See the main README.md for instructions on:
- Building the API service
- Running locally
- Deploying to Kubernetes
- Testing DNS resolution

## Migration from PolyCloud

Conn3ction is a rebranding and evolution of the PolyCloud concept. Key changes:
- Updated naming throughout codebase
- etcd key prefixes changed from `/polycloud/` to `/conn3ction/`
- Simplified API interface
- Enhanced Helm chart flexibility
