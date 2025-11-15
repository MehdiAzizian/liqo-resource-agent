# Changelog

All notable changes to this project will be documented in this file.

## [Phase 1 - Update] - 2025-11-15

### Changed
- **Event-driven controller**: Now watches Nodes and Pods for immediate updates
- Controller responds to changes in < 1 second instead of waiting for 30-second interval
- Added watch handlers for Node and Pod resources
- Hybrid approach: Event-driven + periodic backup reconciliation

### Technical Details
- Added `Watches()` for Node and Pod resources in controller setup
- Implemented `findAdvertisementsForNode()` and `findAdvertisementsForPod()` mapper functions
- Controller now reconciles immediately when cluster state changes

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
