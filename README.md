# Kubernetes Resource Agent

A Kubernetes operator that monitors cluster resources in real-time and publishes resource availability for multi-cluster resource brokerage.

**Master's Thesis Project** - Multi-Cluster Resource Management System

---

## Overview

The Resource Agent automatically monitors a Kubernetes cluster's available resources (CPU, Memory, Storage) and publishes this information to a central Resource Broker for intelligent workload distribution across multiple clusters.

### Key Features

✅ **Real-time Monitoring**
- Tracks CPU, Memory, and Storage resources
- Event-driven updates (<1 second response time)
- Automatic periodic sync (30 seconds)

✅ **Automatic Broker Communication**
- Publishes cluster advertisements to Resource Broker
- Kubernetes-native authentication
- Preserves broker-managed resource locks

✅ **Resource Metrics**
- **Capacity**: Total physical resources
- **Allocatable**: Available to workloads
- **Allocated**: Currently requested
- **Available**: Free for new workloads

---

## Architecture
```
┌─────────────────────────────┐
│   Kubernetes Cluster        │
│                             │
│   Nodes + Pods              │
│         ↓                   │
│   Resource Agent            │
│         ↓                   │
│   Advertisement (CRD)       │
└─────────┬───────────────────┘
          │ HTTPS
          ↓
    ┌──────────────┐
    │    Broker    │
    │   Cluster    │
    └──────────────┘
```

---

## Quick Start

### Prerequisites
- Go 1.23+
- Kubernetes cluster (tested with kind)
- kubectl configured

### Installation
```bash
# Install CRDs
make install

# Run locally
make run

# With broker connection
go run ./cmd/main.go --broker-kubeconfig=/path/to/broker/config
```

### View Resources
```bash
kubectl get advertisements
kubectl describe advertisement cluster-advertisement
```

---

## Example Output
```yaml
apiVersion: rear.fluidos.eu/v1alpha1
kind: Advertisement
metadata:
  name: cluster-advertisement
spec:
  clusterID: "fd32c7d2-7cc6-46e6-80aa-d3d5c835586c"
  resources:
    capacity:
      cpu: "10"
      memory: "8025424Ki"
    allocated:
      cpu: "1050m"
      memory: "418Mi"
    available:
      cpu: "8950m"
      memory: "7597392Ki"
```

---

## Project Structure
```
liqo-resource-agent/
├── api/v1alpha1/          # CRD definitions
├── cmd/main.go            # Entry point
├── internal/
│   ├── controller/        # Advertisement controller
│   ├── metrics/           # Resource collector
│   └── publisher/         # Broker client
└── config/                # Kubernetes manifests
```

---

## Development

### Build
```bash
make build
```

### Generate CRDs
```bash
make manifests
```

### Run Tests
```bash
make test
```

---

## Configuration

### Command-Line Flags

- `--broker-kubeconfig`: Path to broker cluster kubeconfig (optional)
- `--health-probe-bind-address`: Health probe address (default: `:8081`)
- `--metrics-bind-address`: Metrics endpoint (default: `:8080`)

---

## Documentation

- [Phase 1 Report](PHASE1_REPORT.md) - Implementation details
- [Phase 3 Completion](PHASE3_COMPLETE.md) - Communication layer
- [Quick Reference](QUICKREF.md) - Development guide

---

## Related Repository

- [liqo-resource-broker](https://github.com/mehdiazizian/liqo-resource-broker) - Central resource broker

---

## License

Apache License 2.0

## Author

Mehdi Azizian - Master's Thesis Project (2025)