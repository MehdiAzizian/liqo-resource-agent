# Liqo Resource Agent

**Cluster-Side Resource Monitoring and Advertisement Agent**

A Master's thesis project component that runs in each Kubernetes cluster to monitor local resources, publish advertisements to the centralized broker, and receive reservation notifications.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Transport Protocols](#transport-protocols)
- [Certificate Setup](#certificate-setup)
- [Development](#development)
- [Thesis Context](#thesis-context)

---

## 🎯 Overview

The **Liqo Resource Agent** runs in each Kubernetes cluster and:

1. **Monitors Local Resources** (nodes, pods, reservations)
2. **Publishes Advertisements** to the broker about available resources
3. **Receives Reservations** from the broker
4. **Creates Local Instructions** for resource consumption or provisioning

### Key Capabilities

- ✅ Real-time resource monitoring (CPU, Memory, GPU)
- ✅ Protocol-agnostic communication (HTTP or Kubernetes)
- ✅ Automatic cluster ID detection
- ✅ Reserved resource calculation (prevents double-booking)
- ✅ Reservation instruction handling
- ✅ mTLS certificate authentication (HTTP mode)
- ✅ Periodic advertisement updates (clock-synchronized)

---

## 🏗️ Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                  Agent Cluster                          │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  AdvertisementReconciler                         │  │
│  │  - Collects resource metrics every 30s           │  │
│  │  - Publishes to broker                           │  │
│  │  - Updates local Advertisement CRD               │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  MetricsCollector                                │  │
│  │  - Lists nodes (capacity, allocatable)           │  │
│  │  - Lists pods (allocated resources)              │  │
│  │  - Calculates reserved (ProviderInstructions)    │  │
│  │  - Computes Available = Allocatable - Allocated  │  │
│  │                        - Reserved                │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  BrokerCommunicator (Transport Abstraction)      │  │
│  │  ├─ HTTP: POST advertisements, GET reservations  │  │
│  │  └─ Kubernetes: Create CRDs, Watch for updates   │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  ReservationPoller / Watcher                     │  │
│  │  - Polls broker every 30s (HTTP)                 │  │
│  │  - or Watches broker CRDs (Kubernetes)           │  │
│  │  - Creates ReservationInstruction (requester)    │  │
│  │  - Creates ProviderInstruction (provider)        │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Custom Resource Definitions (CRDs)              │  │
│  │  - Advertisement (local cluster state)           │  │
│  │  - ReservationInstruction (use resources)        │  │
│  │  - ProviderInstruction (reserve resources)       │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

```
Local Cluster                     Broker                    Remote Cluster
─────────────                    ────────                   ──────────────

1. Monitor Resources
   ├─ List Nodes
   ├─ List Pods
   └─ List ProviderInstructions
   Calculate Available

2. Publish Advertisement
   ├─ HTTP POST /advertisements   ──────>  Store & process
   └─ or Create CRD                        Run decision engine

3. Poll for Reservations
   <──────  GET /reservations
            (every 30 seconds)

4. Receive Reservation (as Requester)
   Create ReservationInstruction
   "Use remote-cluster for 2CPU/8Gi"

5. Receive Reservation (as Provider)
   Create ProviderInstruction
   "Reserve 2CPU/8Gi for requester"
   (Reduces local Available)
```

---

## ✨ Features

### 1. **Resource Monitoring**

Automatic collection of cluster resources:

**From Nodes:**
- Capacity (total physical resources)
- Allocatable (capacity - system reserved)
- GPU support (nvidia.com/gpu)

**From Pods:**
- Allocated (sum of container requests)
- Includes init containers (max value)
- Includes pod overhead

**From ProviderInstructions:**
- Reserved (resources locked by broker reservations)
- Only counts enforced, non-expired instructions

**Formula:**
```
Available = Allocatable - Allocated - Reserved
```

### 2. **Protocol-Agnostic Communication**

The agent supports multiple transport protocols:

**HTTP REST API (Recommended)**
```bash
./agent --broker-transport=http \
        --broker-url=https://broker.example.com:8443 \
        --broker-cert-path=./certs/agent-cluster-1
```

**Kubernetes CRD-based (Legacy)**
```bash
./agent --broker-kubeconfig=/path/to/broker-kubeconfig
```

### 3. **Clock-Synchronized Updates**

Advertisements are published at synchronized clock times:
- Default interval: 30 seconds
- Updates at: 10:00:00, 10:00:30, 10:01:00, etc.
- Ensures all agents publish simultaneously
- Easier for broker to process batches

### 4. **Reservation Handling**

**As Requester (needs resources):**
- Receives reservation from broker
- Creates local `ReservationInstruction` CRD
- Contains: target cluster, CPU, memory, expiration
- Triggers local automation (Liqo peering in production)

**As Provider (provides resources):**
- Receives reservation from broker
- Creates local `ProviderInstruction` CRD
- Marks resources as "reserved"
- Reduces available resources in future advertisements

### 5. **Critical: Reserved Field Preservation**

When publishing advertisements via HTTP:
1. Fetches existing advertisement from broker
2. Preserves the `reserved` field (broker-managed)
3. Includes preserved value in new publication
4. **Prevents race conditions** in resource locking

---

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or later
- Kubernetes cluster (where agent will run)
- kubectl configured to access local cluster
- Access to broker (HTTP URL or kubeconfig)
- cert-manager installed in the cluster (for HTTP transport)

### Installation

1. **Clone the repository**

```bash
git clone <your-repo-url>
cd liqo-resource-agent
```

2. **Install dependencies**

```bash
go mod download
```

3. **Build the agent**

```bash
make build
# or
go build -o bin/agent cmd/main.go
```

4. **Install CRDs**

```bash
make install
# or
kubectl apply -f config/crd/bases/
```

### Running the Agent

The agent uses **ONE transport protocol** to communicate with the broker.
The transport must match the broker's `--broker-interface` setting.

#### Option 1: HTTP Transport (Recommended)

```bash
# First, setup certificates with cert-manager (see Certificate Setup)
kubectl apply -k config/certmanager/

# Start agent with HTTP transport
./bin/agent \
  --broker-transport=http \
  --broker-url=https://broker.example.com:8443 \
  --broker-cert-path=/etc/agent/certs \
  --cluster-id=cluster-1 \
  --health-probe-bind-address=:8081
```

Broker must use: `--broker-interface=http`

#### Option 2: Kubernetes Transport (Legacy)

```bash
./bin/agent \
  --broker-transport=kubernetes \
  --broker-kubeconfig=/path/to/broker-kubeconfig \
  --health-probe-bind-address=:8081
```

Broker must use: `--broker-interface=kubernetes`

**Note:** The cluster ID is automatically detected from `kube-system` namespace UID, or can be overridden with `--cluster-id`. When using HTTP transport with cert-manager, the cluster ID comes from the certificate CN field.

---

## ⚙️ Configuration

### Command-Line Flags

#### Transport Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--broker-transport` | `""` | Transport protocol: `http`, `kubernetes`, or empty (disabled) |
| `--broker-url` | `""` | Broker URL for HTTP transport |
| `--broker-cert-path` | `""` | Client certificate path for HTTP transport |
| `--broker-kubeconfig` | `""` | Kubeconfig for Kubernetes transport (legacy) |
| `--broker-namespace` | `default` | Namespace for broker CRDs |

#### Agent Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--cluster-id` | auto | Override cluster ID (defaults to kube-system UID) |
| `--advertisement-name` | `cluster-advertisement` | Advertisement resource name |
| `--advertisement-namespace` | `default` | Advertisement namespace |
| `--instruction-namespace` | same as advertisement | Namespace for Instruction CRDs |
| `--advertisement-requeue-interval` | `30s` | Update interval |

#### Standard Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-bind-address` | `:8080` | Address for metrics endpoint |
| `--health-probe-bind-address` | `:8081` | Address for health probes |
| `--leader-elect` | `false` | Enable leader election |

### Environment Variables

- `KUBECONFIG`: Path to local cluster kubeconfig
- `POD_NAMESPACE`: Namespace where agent is running

---

## 🌐 Transport Protocols

### HTTP Transport (Recommended)

**Advantages:**
- ✅ No kubeconfig sharing required
- ✅ Certificate-based authentication (mTLS)
- ✅ Works with non-Kubernetes brokers
- ✅ Standard HTTPS security
- ✅ Easy firewall configuration

**How it works:**
1. Agent publishes advertisements via `POST /api/v1/advertisements`
2. Agent polls reservations via `GET /api/v1/reservations` every 30s
3. Cluster ID extracted from certificate CN
4. All traffic encrypted with TLS 1.2+

**Configuration:**
```bash
./agent --broker-transport=http \
        --broker-url=https://broker.example.com:8443 \
        --broker-cert-path=./certs/agent-cluster-1 \
        --cluster-id=cluster-1
```

**Certificate Requirements:**
- Client certificate signed by broker's CA
- Certificate CN must match cluster ID
- Example: `CN=cluster-1, O=LiqoResourceBroker`

---

### Kubernetes CRD Transport (Legacy)

**Advantages:**
- ✅ Native Kubernetes approach
- ✅ Watch-based notifications (push, not poll)
- ✅ Uses existing RBAC

**Disadvantages:**
- ❌ Requires kubeconfig sharing
- ❌ Tightly coupled to Kubernetes
- ❌ Complex RBAC setup
- ❌ Not protocol-agnostic

**How it works:**
1. Agent creates `ClusterAdvertisement` CRDs in broker cluster
2. Agent watches broker's `Reservation` CRDs
3. Uses Kubernetes API for all communication

**Configuration:**
```bash
./agent --broker-kubeconfig=/path/to/broker-kubeconfig
```

**Note:** Legacy mode is maintained for backward compatibility and thesis comparison. New deployments should use HTTP transport.

---

## 🔐 Certificate Setup (cert-manager)

All certificates are managed by **cert-manager**, which automatically creates and renews them.

### Prerequisites

```bash
# Install cert-manager (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

### For Agents in the Same Cluster as Broker

```bash
kubectl apply -k config/certmanager/
```

### For Agents in Different Clusters

```bash
# 1. Export CA secret from broker cluster
kubectl get secret liqo-broker-ca-secret -n cert-manager -o yaml > ca-secret.yaml

# 2. Import CA secret to agent cluster
kubectl create namespace cert-manager  # if not exists
kubectl apply -f ca-secret.yaml

# 3. Edit agent-certificate.yaml to set the correct commonName (cluster ID)
# 4. Apply agent certificate configuration
kubectl apply -k config/certmanager/
```

### How It Works

```
cert-manager watches Certificate resources
  → Generates key pair
  → Signs with shared CA
  → Stores in Kubernetes Secret
  → Auto-renews before expiry
  → Pod mounts Secret as volume
```

**Important:** The certificate's `commonName` (CN) field IS the cluster identity. The broker extracts the cluster ID from this field during mTLS authentication.

### Adding a New Agent Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-cluster-N-cert
  namespace: default
spec:
  secretName: agent-cluster-N-tls
  duration: 8760h
  renewBefore: 720h
  commonName: cluster-N              # THIS is the cluster ID
  usages:
    - client auth
  issuerRef:
    name: liqo-broker-ca-issuer
    kind: ClusterIssuer
```

### Mounting in Agent Pod

```yaml
volumes:
- name: certs
  secret:
    secretName: agent-client-tls
containers:
- name: agent
  volumeMounts:
  - name: certs
    mountPath: /etc/agent/certs      # --broker-cert-path points here
    readOnly: true
```

---

## 🛠️ Development

### Project Structure

```
liqo-resource-agent/
├── api/v1alpha1/              # CRD definitions
│   ├── advertisement_types.go
│   ├── reservationinstruction_types.go
│   └── providerinstruction_types.go
├── cmd/main.go                # Agent entry point
├── internal/
│   ├── controller/            # Kubernetes controllers
│   │   ├── advertisement_controller.go
│   │   ├── reservationinstruction_controller.go
│   │   └── providerinstruction_controller.go
│   ├── metrics/
│   │   └── collector.go       # Resource monitoring
│   ├── publisher/             # LEGACY Kubernetes transport
│   │   ├── broker_client.go
│   │   └── reservation_watcher.go
│   └── transport/             # Protocol abstraction
│       ├── interface.go       # BrokerCommunicator interface
│       ├── factory.go         # Transport selection
│       ├── dto/               # Data transfer objects
│       └── http/              # HTTP implementation
│           ├── client.go
│           └── poller.go
└── config/
    ├── crd/bases/             # CRD manifests
    └── certmanager/           # cert-manager certificate configs
```

### Building

```bash
# Development build
make build

# Production build
make build-prod

# Run tests
make test

# Install CRDs
make install

# Deploy to cluster
make deploy
```

### Running Locally

```bash
# Against local cluster
export KUBECONFIG=~/.kube/config
./bin/agent --broker-transport=http \
            --broker-url=https://localhost:8443 \
            --broker-cert-path=./certs/agent-cluster-1 \
            --cluster-id=test-cluster
```

---

## 🎓 Thesis Context

### Research Problem

**How to decouple cluster resource monitoring from broker communication protocol in multi-cluster environments?**

### Solution Architecture

The agent implements a **transport abstraction layer** that separates business logic from communication:

**Before (Tightly Coupled):**
```
Metrics Collection → Kubernetes API → Broker's CRDs
```

**After (Protocol-Agnostic):**
```
Metrics Collection → Transport Interface → [HTTP | Kubernetes | MQTT | ...]
```

### Key Contributions

1. **Protocol Independence**
   - Communication via `BrokerCommunicator` interface
   - Business logic independent of transport
   - Easy to add new protocols

2. **HTTP Transport Implementation**
   - mTLS certificate authentication
   - RESTful API calls
   - Polling-based reservation notifications
   - Preserved Reserved field logic

3. **Resource Monitoring**
   - Three-way calculation (Nodes, Pods, Instructions)
   - GPU support
   - Reserved resource tracking
   - Clock-synchronized updates

4. **Backward Compatibility**
   - Legacy Kubernetes transport still works
   - Gradual migration path
   - No breaking changes for existing deployments

### Comparison: Before vs After

| Aspect | Before (Kubernetes Only) | After (Protocol-Agnostic) |
|--------|-------------------------|---------------------------|
| **Communication** | Kubernetes CRDs | Interface-based (HTTP, K8s, etc.) |
| **Authentication** | Kubeconfig | mTLS certificates or kubeconfig |
| **Notifications** | Watch (push) | Polling (HTTP) or Watch (K8s) |
| **Extensibility** | Kubernetes-only | Any protocol implementing interface |
| **Deployment** | Complex RBAC | Standard HTTPS or K8s |
| **Coupling** | Tight | Loose |

---

## 📚 API Reference

### BrokerCommunicator Interface

```go
type BrokerCommunicator interface {
    // Publish advertisement to broker
    PublishAdvertisement(ctx context.Context, adv *AdvertisementDTO) error

    // Fetch reservations for this cluster
    FetchReservations(ctx context.Context, clusterID string, role Role) ([]*ReservationDTO, error)

    // Health check
    Ping(ctx context.Context) error

    // Cleanup
    Close() error
}
```

### AdvertisementDTO Structure

```go
type AdvertisementDTO struct {
    ClusterID   string
    ClusterName string
    Resources   ResourceMetricsDTO
    Timestamp   time.Time
}

type ResourceMetricsDTO struct {
    Capacity    ResourceQuantitiesDTO
    Allocatable ResourceQuantitiesDTO
    Allocated   ResourceQuantitiesDTO
    Reserved    *ResourceQuantitiesDTO  // Broker-managed
    Available   ResourceQuantitiesDTO
}
```

---

## 🐛 Troubleshooting

### Agent Not Publishing Advertisements

**Check:**
1. Broker transport configured: `--broker-transport=http`
2. Broker URL reachable: `curl -k https://broker.example.com:8443/healthz`
3. Certificate valid: `openssl verify -CAfile ca.crt tls.crt`
4. Logs: `kubectl logs <agent-pod> -n rear-agent`

### Certificate Authentication Failing

**Check:**
1. Certificate CN matches cluster ID
2. Certificate signed by broker's CA
3. Certificate not expired: `openssl x509 -in tls.crt -noout -dates`
4. CA certificate matches broker's CA

### Reservations Not Received

**Check:**
1. Polling enabled (HTTP transport)
2. Reservation phase is "Reserved"
3. ClusterID matches (requester or provider)
4. Logs show polling activity


