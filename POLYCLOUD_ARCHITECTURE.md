# PolyCloud Architecture

## Overview

PolyCloud is a DNS-based system for region-aware service discovery and routing in multi-cloud and hybrid cloud environments. It provides a unified DNS layer that can route requests based on geographic regions, cloud providers, and service availability.

## Core Components

### 1. DNS API Service

A lightweight Go-based REST API that manages DNS records for the PolyCloud system.

**Features:**
- CRUD operations for DNS A records
- In-memory storage (default) or etcd-backed persistence
- RESTful JSON API
- Health check endpoint
- Graceful shutdown

**Architecture:**
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP (JSON)
       v
┌─────────────┐
│  DNS API    │──────┐
│  (Go HTTP)  │      │ Optional
└──────┬──────┘      │
       │             v
       │      ┌─────────────┐
       │      │    etcd     │
       │      │  (KV Store) │
       │      └─────────────┘
       │
       v
┌─────────────┐
│  MemStore   │
│ (In-Memory) │
└─────────────┘
```

### 2. CoreDNS with etcd Backend

CoreDNS serves as the authoritative DNS server for the `example.local` zone (configurable).

**Deployment Modes:**
1. **Development Mode (ConfigMap-based):** Uses static zone files from Kubernetes ConfigMaps
2. **Production Mode (etcd-based):** Uses etcd as a dynamic backend for DNS records

**Integration Flow:**
```
┌──────────────┐
│DNS Clients   │
│(applications)│
└──────┬───────┘
       │ DNS Query
       v
┌──────────────┐       ┌─────────────┐
│   CoreDNS    │──────>│    etcd     │
│              │<──────│             │
└──────────────┘       └─────────────┘
       │                      ^
       │                      │
       └──────────────────────┘
          Zone: example.local
```

### 3. Storage Layer

**In-Memory Store:**
- Fast, no external dependencies
- Suitable for development and testing
- Data lost on restart

**etcd Store:**
- Distributed, persistent storage
- High availability support
- Suitable for production
- Records stored at `/polycloud/records/{fqdn}`

## Data Model

### DNS Record
```json
{
  "name": "app.example.local",
  "type": "A",
  "values": ["10.1.2.3"],
  "ttl": 60
}
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/healthz` | Health check |
| GET | `/records` | List all records |
| POST | `/records` | Create a new record |
| GET | `/records/{name}` | Get a specific record |
| PUT | `/records/{name}` | Update a record |
| DELETE | `/records/{name}` | Delete a record |

## Deployment Architecture

### Kubernetes Components

```
┌─────────────────────────────────────────┐
│         Kubernetes Cluster              │
│                                         │
│  ┌──────────────┐   ┌──────────────┐  │
│  │  CoreDNS     │   │  DNS API     │  │
│  │  Deployment  │   │  Deployment  │  │
│  └──────┬───────┘   └──────┬───────┘  │
│         │                   │           │
│         │   ┌───────────────┘           │
│         │   │                           │
│         v   v                           │
│  ┌──────────────┐                      │
│  │     etcd     │                      │
│  │ StatefulSet  │                      │
│  └──────────────┘                      │
│                                         │
└─────────────────────────────────────────┘
```

## Future Enhancements

1. **Region Manager:** Intelligent routing based on client location
2. **Multi-region Support:** Active-active DNS across regions
3. **RBAC:** Role-based access control for API
4. **Monitoring:** Prometheus metrics for DNS queries and API requests
5. **DNS Record Types:** Support for CNAME, SRV, TXT records
6. **TTL Management:** Dynamic TTL based on health checks
7. **Health Checks:** Integration with service health endpoints
8. **Caching:** Intelligent caching strategies for performance

## Security Considerations

1. **Authentication:** Currently none - add API keys or OAuth2
2. **Authorization:** RBAC for record management
3. **TLS:** Enable TLS for API and etcd communication
4. **Network Policies:** Restrict access between components
5. **Secret Management:** Use Kubernetes secrets for sensitive data

## Getting Started

See the main [README.md](README.md) for quickstart instructions.
