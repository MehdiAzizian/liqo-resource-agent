# Liqo Resource Agent (liqo-ra)

Kubernetes controller that automatically advertises cluster resources for the FLUIDOS resource brokerage system.

## Quick Start

### Prerequisites
- Go 1.25+
- kubectl
- Docker
- kind (for local testing)

### Installation

1. **Clone and navigate:**
```bash
   cd liqo-resource-agent
```

2. **Install CRDs:**
```bash
   make install
```

3. **Run controller locally:**
```bash
   make run
```

4. **Create an Advertisement:**
```bash
   kubectl apply -f config/samples/rear_v1alpha1_advertisement.yaml
```

5. **View resources:**
```bash
   kubectl get advertisements
```

## What It Does

The Resource Agent **watches** your Kubernetes cluster and immediately reacts to changes, maintaining an up-to-date Advertisement of available resources:

- **Event-Driven**: Responds instantly when pods are created/deleted or nodes change
- **Automatic Updates**: No manual intervention needed
- **Comprehensive Metrics**:
  - **Capacity**: Total hardware resources
  - **Allocatable**: Resources available to pods
  - **Allocated**: Resources requested by running pods  
  - **Available**: Resources still free for scheduling

The controller watches:
- ✅ Pod lifecycle events (create, delete, status changes)
- ✅ Node changes (add, remove, capacity updates)
- ✅ Resource request modifications

**Response Time**: Immediate (< 1 second) + periodic backup every 30 seconds

## Example Output
```bash
$ kubectl get advertisements
NAME                    CLUSTERID       ALLOCATABLE-CPU   AVAILABLE-CPU   ALLOCATABLE-MEM   AVAILABLE-MEM   PUBLISHED   AGE
cluster-advertisement   fd32c7d2-...    10                9050m           8025424Ki         7728464Ki       true        2m
```

## Architecture
```
┌─────────────────────────────────────┐
│   Kubernetes Cluster (kind)         │
│                                      │
│  ┌────────┐  ┌────────┐  ┌────────┐│
│  │ Node 1 │  │ Node 2 │  │ Pod A  ││
│  └────────┘  └────────┘  └────────┘│
│       ▲           ▲           ▲     │
│       └───────────┴───────────┘     │
│                   │                 │
└───────────────────┼─────────────────┘
                    │
            ┌───────▼────────┐
            │ Metrics        │
            │ Collector      │
            └───────┬────────┘
                    │
            ┌───────▼────────┐
            │ Advertisement  │
            │ Controller     │
            └───────┬────────┘
                    │
            ┌───────▼────────┐
            │ Advertisement  │
            │ CR (stored in  │
            │ etcd)          │
            └────────────────┘
```

## Project Structure
```
.
├── api/v1alpha1/              # CRD definitions
├── internal/
│   ├── controller/            # Reconciliation logic
│   └── metrics/               # Resource collection
├── config/                    # Kubernetes manifests
│   ├── crd/                   # Generated CRDs
│   ├── rbac/                  # RBAC rules
│   └── samples/               # Example CRs
└── cmd/                       # Main entrypoint
```

## Development

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Generate Manifests
```bash
make manifests
```

### Deploy to Cluster
```bash
make deploy
```

## Configuration

The controller requires these RBAC permissions:
- `advertisements.rear.fluidos.eu` (all operations)
- `nodes` (list, watch)
- `pods` (list, watch)
- `namespaces` (get)

## Documentation

- [Phase 1 Report](PHASE1_REPORT.md) - Detailed technical explanation
- [API Reference](api/v1alpha1/advertisement_types.go) - CRD schema

## License

Apache 2.0

## Project Info

- **Domain**: fluidos.eu
- **API Group**: rear.fluidos.eu
- **Version**: v1alpha1
- **Kind**: Advertisement
