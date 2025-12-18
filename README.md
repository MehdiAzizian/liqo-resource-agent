# Kubernetes Resource Agent

A Kubernetes operator that monitors cluster resources in real-time, publishes availability to a central broker, and **receives reservation instructions** for multi-cluster resource brokerage.

**Master's Thesis Project** - Multi-Cluster Resource Management System

---

## Overview

The Resource Agent is the "local" component running in each participating cluster. It has two main responsibilities:
1.  **Publish:** Automatically monitors the local cluster's available resources (CPU, Memory) and pushes this data to the central Resource Broker.
2.  **Watch:** Listens for "Reserved" events from the Broker directed at this cluster, enabling the local manager to initiate peering or workload offloading.

### Key Features

✅ **Real-time Monitoring**
- Tracks CPU and Memory resources
- Aggregates node capacity and current pod allocations
- Calculates true "Available" resources

✅ **Broker Publishing**
- Publishes `ClusterAdvertisement` CRs to the central Broker
- Updates automatically on resource changes or periodically (30s)

✅ **Feedback Loop (New)**
- **Reservation Watcher:** Connects to the Broker and watches for `Reservation` objects.
- **Requester Awareness:** Filters reservations to only react when `spec.requesterID` matches the local cluster ID.
- **Provider Awareness:** Also reacts when `spec.targetClusterID` equals the local cluster to signal that this cluster must keep capacity aside.
- **Notifications:** Persists a `ReservationInstruction` (requester view) and a `ProviderInstruction` (provider view) CR in the local cluster so automation can react to either responsibility without scraping stdout.

---

## Architecture
┌─────────────────────────────┐ │ Kubernetes Cluster │ │ (e.g., "Rome") │ │ │ │ Resource Agent │ │ ┌─────────────────┐ │ │ │ MetricsCollector│ │ │ └────────┬────────┘ │ │ │ 1. Publish │ │ ▼ │ │ ┌─────────────────┐ │ 2. Watch for Reservations │ │ BrokerClient │◄───┼────────────────────────┐ │ └────────┬────────┘ │ (RequesterID="Rome")│ └───────────────┼─────────────┘ │ │ HTTPS │ ▼ │ ┌──────────────┐ │ │ Broker │──────────────────────────────┘ │ Cluster │ └──────────────┘


---

## Quick Start

### Prerequisites
- Go 1.23+
- Kubernetes cluster
- `kubectl` configured
- Access to the **Broker Cluster** (kubeconfig)

### Installation

1. **Install CRDs**
   ```bash
   make install
   ```
Run the Agent You must provide the kubeconfig for the central Broker so the Agent can publish advertisements and watch for reservations.
```bash
go run ./cmd/main.go \
  --broker-kubeconfig=/path/to/broker/kubeconfig \
  --advertisement-name=cluster-advertisement \
  --advertisement-namespace=default
```
> Optional flags: `--cluster-id` (override kube-system UID), `--broker-namespace` (if broker CRDs live outside `default`), `--instruction-namespace` to store `ReservationInstruction` CRs elsewhere.
Logs
You will see logs indicating the agent is watching:

INFO    setup    Broker client initialized successfully
INFO    reservation-watcher    Starting reservation watcher    {"clusterID": "rome-cluster"}
When a reservation is fulfilled by the Broker:

INFO    reservation-watcher    !!! RESERVATION FULFILLED !!!    {"requester": "rome", "target": "paris", ...}
INFO    reservation-watcher    Manager Notification: Use paris for 4 CPU, 8Gi Memory
INFO    reservation-watcher    Created ReservationInstruction   {"name": "rome-request", "namespace": "default"}
INFO    reservation-watcher    Provider reservation received    {"reservation": "rome-request", "requester": "rome"}
INFO    providerinstruction    Enforcing provider instruction   {"reservation": "rome-request", "requester": "rome"}
Project Structure
liqo-resource-agent/
├── api/v1alpha1/          # CRD definitions
├── cmd/main.go            # Entry point (Manager + Watcher)
├── internal/
│   ├── controller/        # Advertisement controller
│   ├── metrics/           # Resource collector
│   └── publisher/         # Broker client & Reservation Watcher
└── config/                # Kubernetes manifests
License
Apache License 2.0

Author
Mehdi Azizian - Master's Thesis Project (2025)
