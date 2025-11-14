# Phase 1 Report: Resource Agent (RA) Implementation

## Overview
This phase implements a Kubernetes controller that acts as a Resource Agent (RA) for the FLUIDOS resource brokerage system. The agent automatically collects, tracks, and advertises cluster resources.

---

## Architecture

### Components Built

1. **Advertisement CRD (Custom Resource Definition)**
   - Location: `api/v1alpha1/advertisement_types.go`
   - Purpose: Defines the schema for advertising cluster resources
   
2. **Metrics Collector**
   - Location: `internal/metrics/collector.go`
   - Purpose: Collects resource information from Kubernetes cluster
   
3. **Advertisement Controller**
   - Location: `internal/controller/advertisement_controller.go`
   - Purpose: Reconciliation logic that keeps advertisements up-to-date

---

## How It Works

### 1. Advertisement Custom Resource

The Advertisement CRD represents a snapshot of cluster resources at a point in time.

**Key Fields:**
```yaml
spec:
  clusterID: "unique-cluster-identifier"
  resources:
    capacity:      # Total physical resources
      cpu: "10"
      memory: "8Gi"
    allocatable:   # Resources available to pods (capacity - system reserved)
      cpu: "10"
      memory: "7.8Gi"
    allocated:     # Resources requested by running pods
      cpu: "950m"
      memory: "290Mi"
    available:     # Resources still schedulable (allocatable - allocated)
      cpu: "9050m"
      memory: "7.5Gi"
  timestamp: "2025-11-14T22:15:30Z"

status:
  phase: "Active"
  published: true
  message: "Advertisement updated successfully"
  lastUpdateTime: "2025-11-14T22:15:30Z"
```

**Resource Metrics Explained:**

- **Capacity**: Total hardware resources in the cluster (sum of all nodes)
- **Allocatable**: Capacity minus Kubernetes system reservations (kubelet, OS, etc.)
- **Allocated**: Sum of resource *requests* from all running/pending pods
- **Available**: Allocatable minus Allocated = what can still be scheduled
- **Used** (future): Actual consumption from metrics-server (not implemented yet)

### 2. Metrics Collector (`internal/metrics/collector.go`)

**Purpose**: Queries the Kubernetes API to gather resource information.

**Key Functions:**
```go
// CollectClusterResources() - Main collection function
// Returns: ResourceMetrics with all breakdowns
func (c *Collector) CollectClusterResources(ctx) (*ResourceMetrics, error)

// calculateAllocatedResources() - Sums pod requests
// Iterates through all pods and adds up their resource requests
func (c *Collector) calculateAllocatedResources(ctx) (*ResourceQuantities, error)

// GetClusterID() - Returns unique cluster identifier
// Uses kube-system namespace UID as stable cluster ID
func (c *Collector) GetClusterID(ctx) (string, error)
```

**How Resource Collection Works:**

1. **List all nodes** in the cluster
2. **For each ready node**, aggregate:
   - `node.Status.Capacity` → total Capacity
   - `node.Status.Allocatable` → total Allocatable
3. **List all pods** in the cluster
4. **For each running/pending pod**, sum up:
   - `container.Resources.Requests` → total Allocated
5. **Calculate Available**:
```
   Available = Allocatable - Allocated
```

**GPU Support:**
- Detects `nvidia.com/gpu` resources on nodes
- Aggregates GPU capacity, allocatable, and allocated amounts

### 3. Advertisement Controller (`internal/controller/advertisement_controller.go`)

**Purpose**: Kubernetes reconciliation loop that keeps Advertisement resources updated.

**Reconciliation Logic:**
```
Every 30 seconds or when Advertisement changes:
  1. Fetch the Advertisement resource
  2. Collect current cluster metrics
  3. Get cluster ID
  4. Update Advertisement.Spec with new data
  5. Update Advertisement.Status (phase, published, timestamp)
  6. Requeue after 30 seconds
```

**Controller Flow:**
```
User creates Advertisement CR
         ↓
Controller detects new resource
         ↓
Reconcile() is called
         ↓
Collect metrics from cluster
         ↓
Update Advertisement.Spec with metrics
         ↓
Update Advertisement.Status = "Active", Published = true
         ↓
Wait 30 seconds
         ↓
Reconcile() called again (periodic update)
         ↓
(repeat)
```

**RBAC Permissions:**

The controller needs these permissions (defined via kubebuilder annotations):
```yaml
# For Advertisement CRD
- advertisements.rear.fluidos.eu (all verbs)
- advertisements/status (get, update, patch)

# For metrics collection
- nodes (get, list, watch)
- pods (get, list, watch)
- namespaces (get, list, watch)
```

---

## Project Structure
```
tesi2/
├── api/v1alpha1/
│   ├── advertisement_types.go      # CRD definition
│   └── groupversion_info.go        # API group metadata
├── internal/
│   ├── controller/
│   │   └── advertisement_controller.go  # Reconciliation logic
│   └── metrics/
│       └── collector.go            # Resource collection
├── config/
│   ├── crd/bases/                  # Generated CRD YAML
│   ├── rbac/                       # Generated RBAC rules
│   ├── manager/                    # Controller deployment config
│   └── samples/
│       └── rear_v1alpha1_advertisement.yaml  # Sample CR
├── cmd/
│   └── main.go                     # Controller entrypoint
├── Makefile                        # Build and deployment tasks
├── go.mod                          # Go dependencies
└── go.sum
```

---

## Key Technologies & Concepts

### Kubebuilder
- Framework for building Kubernetes controllers
- Generates boilerplate code, CRDs, RBAC, Makefiles
- Uses controller-runtime library

### Controller-Runtime
- Library for building Kubernetes controllers
- Provides:
  - Manager: orchestrates controllers
  - Client: interacts with Kubernetes API
  - Reconciler: implements business logic

### Custom Resource Definition (CRD)
- Extends Kubernetes API with custom resources
- Advertisement is a custom resource type
- Kubebuilder markers (comments starting with `+kubebuilder:`) generate CRD YAML

### Reconciliation Loop
- Core pattern in Kubernetes controllers
- Continuously ensures desired state matches actual state
- Level-triggered (not edge-triggered): runs periodically and on changes

---

## Testing & Validation

### Local Setup
- **Cluster**: kind (Kubernetes in Docker)
- **Cluster Name**: liqo-test
- **Controller Mode**: Running locally (`make run`)

### Verification Commands
```bash
# Check CRD is installed
kubectl get crds | grep advertisements

# List advertisements
kubectl get advertisements

# View detailed advertisement
kubectl describe advertisement cluster-advertisement

# View YAML output
kubectl get advertisement cluster-advertisement -o yaml

# Watch for updates
kubectl get advertisements -w
```

### Expected Behavior
- Advertisement created → Controller detects it
- Within seconds: metrics populated automatically
- Every 30 seconds: metrics refresh
- Status shows: Phase=Active, Published=true

---

## What We Achieved

✅ **Automatic resource discovery**: No manual input needed
✅ **Comprehensive metrics**: Capacity, Allocatable, Allocated, Available
✅ **Continuous updates**: Fresh data every 30 seconds
✅ **REAR-compatible structure**: Ready for broker integration
✅ **GPU support**: Detects and reports GPU resources
✅ **Cluster identification**: Stable cluster ID from kube-system UID
✅ **Status tracking**: Observable state via kubectl

---

## Next Steps (Phase 2)

1. Build Resource Broker (RB) controller
2. Enable multiple clusters to register with broker
3. Implement aggregation of advertisements from multiple RAs
4. Add decision logic for resource selection
5. Create reservation request mechanism

---

## Limitations & Future Enhancements

### Current Limitations
1. **No "Used" metrics**: Requires metrics-server integration
2. **No broker communication**: Advertisements stay local
3. **No authentication**: Trust between RA and RB not implemented
4. **No cost modeling**: Cost fields exist but aren't calculated
5. **Local only**: Controller runs outside cluster (dev mode)

### Potential Enhancements
1. Integrate metrics-server for real-time usage data
2. Add HTTP/gRPC endpoint to push advertisements to broker
3. Implement TLS authentication
4. Add cost calculation based on cloud provider pricing
5. Deploy controller as in-cluster Deployment
6. Add webhook for validation/defaulting
7. Implement finalizers for cleanup logic

---

## Code Quality Notes

### Good Practices Implemented
- Structured logging with contextual information
- Error handling with proper error wrapping
- Resource cleanup (no leaks)
- RBAC markers for security
- Status subresource for observability
- Kubebuilder markers for maintainability

### Areas for Improvement
- Add unit tests for metrics collector
- Add integration tests for controller
- Implement retry logic with exponential backoff
- Add metrics/prometheus instrumentation
- Better handling of concurrent updates (optimistic locking)

---

## Glossary

- **CRD**: Custom Resource Definition - extends Kubernetes API
- **CR**: Custom Resource - instance of a CRD
- **RA**: Resource Agent - cluster-side component
- **RB**: Resource Broker - central aggregation component
- **REAR**: Resource Exchange and Advertisement for the Continuum (FLUIDOS protocol)
- **Reconciliation**: Process of ensuring desired state matches actual state
- **Controller**: Component that watches resources and takes action
- **Operator**: Controller that manages complex applications (RA is a simple operator)

---

*Generated: November 14, 2025*
*Phase: 1 - Resource Agent (RA)*
*Project: Automatic Cloud Resource Brokerage for Kubernetes*
