# Changelog

All notable changes to this project will be documented in this file.

## [Phase 1] - 2025-11-14

### Added
- Initial Resource Agent implementation
- Advertisement CRD with comprehensive resource metrics
- Metrics collector for node and pod resources
- Advertisement controller with 30-second reconciliation loop
- Support for CPU, Memory, and GPU resources
- Five-tier resource breakdown: Capacity, Allocatable, Allocated, Used (placeholder), Available
- Cluster identification using kube-system namespace UID
- Status tracking with phase and publication state
- Enhanced kubectl display columns for resource overview
- Local testing with kind cluster
- Project documentation (README, Phase 1 Report)

### Technical Details
- Built with Kubebuilder v4.9.0
- Go 1.25.3
- Kubernetes 1.34.0
- Controller-runtime v0.22.1
- Domain: fluidos.eu
- API Group: rear.fluidos.eu/v1alpha1
