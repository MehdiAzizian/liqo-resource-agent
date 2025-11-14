# Quick Reference - Resource Agent

## Common Commands

### Development
```bash
# Build the controller
make build

# Run locally (connects to current kubectl context)
make run

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Update dependencies
go mod tidy

# Generate/update CRDs and RBAC
make manifests
```

### Deployment
```bash
# Install CRDs to cluster
make install

# Uninstall CRDs
make uninstall

# Deploy controller to cluster
make deploy

# Undeploy from cluster
make undeploy
```

### Testing
```bash
# Create test cluster
kind create cluster --name liqo-test

# Delete test cluster
kind delete cluster --name liqo-test

# Apply sample Advertisement
kubectl apply -f config/samples/rear_v1alpha1_advertisement.yaml

# Watch advertisements
kubectl get advertisements -w

# Describe advertisement
kubectl describe advertisement cluster-advertisement

# Get YAML
kubectl get advertisement cluster-advertisement -o yaml

# Delete advertisement
kubectl delete advertisement cluster-advertisement
```

### Useful kubectl Commands
```bash
# List all CRDs
kubectl get crds

# Check if Advertisement CRD exists
kubectl get crd advertisements.rear.fluidos.eu

# View CRD definition
kubectl get crd advertisements.rear.fluidos.eu -o yaml

# List all advertisements across all namespaces
kubectl get advertisements -A

# Get events related to advertisements
kubectl get events --field-selector involvedObject.kind=Advertisement
```

## File Locations

| Component | Location |
|-----------|----------|
| Advertisement CRD | `api/v1alpha1/advertisement_types.go` |
| Controller Logic | `internal/controller/advertisement_controller.go` |
| Metrics Collector | `internal/metrics/collector.go` |
| Generated CRDs | `config/crd/bases/` |
| Sample CR | `config/samples/rear_v1alpha1_advertisement.yaml` |
| Main Entry | `cmd/main.go` |

## Key Concepts

### Resource Tiers
1. **Capacity** = Total hardware
2. **Allocatable** = Capacity - System Reserved
3. **Allocated** = Sum of pod requests
4. **Available** = Allocatable - Allocated
5. **Used** = Actual consumption (not implemented)

### Reconciliation Interval
- Default: 30 seconds
- Change in: `advertisement_controller.go` line ~90

### Cluster ID
- Source: kube-system namespace UID
- Stable across controller restarts

## Troubleshooting

### Controller won't start
```bash
# Check if CRDs are installed
kubectl get crds | grep advertisement

# Reinstall CRDs
make install
```

### No metrics showing
```bash
# Check if nodes are ready
kubectl get nodes

# Check if pods exist
kubectl get pods -A

# View controller logs
# (logs appear in terminal where 'make run' is running)
```

### Advertisement not updating
```bash
# Check controller logs for errors
# Verify advertisement exists
kubectl get advertisements

# Delete and recreate
kubectl delete advertisement cluster-advertisement
kubectl apply -f config/samples/rear_v1alpha1_advertisement.yaml
```

## Environment

- **Go Version**: 1.25.3
- **Kubebuilder**: 4.9.0
- **Kubernetes**: 1.34.0
- **Controller Runtime**: 0.22.1
- **Domain**: fluidos.eu
- **API Group**: rear.fluidos.eu
- **API Version**: v1alpha1

