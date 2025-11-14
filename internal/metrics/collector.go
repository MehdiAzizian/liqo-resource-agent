package metrics

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
)

// Collector collects resource metrics from the cluster
type Collector struct {
	Client client.Client
}

// CollectClusterResources collects detailed resource information from all nodes
func (c *Collector) CollectClusterResources(ctx context.Context) (*rearv1alpha1.ResourceMetrics, error) {
	nodeList := &corev1.NodeList{}
	if err := c.Client.List(ctx, nodeList); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodeList.Items) == 0 {
		return nil, fmt.Errorf("no nodes found in cluster")
	}

	// Initialize totals
	capacity := &rearv1alpha1.ResourceQuantities{
		CPU:    *resource.NewQuantity(0, resource.DecimalSI),
		Memory: *resource.NewQuantity(0, resource.BinarySI),
	}
	allocatable := &rearv1alpha1.ResourceQuantities{
		CPU:    *resource.NewQuantity(0, resource.DecimalSI),
		Memory: *resource.NewQuantity(0, resource.BinarySI),
	}

	var capacityGPU, allocatableGPU resource.Quantity
	hasGPU := false

	// Aggregate capacity and allocatable from all ready nodes
	for _, node := range nodeList.Items {
		if !isNodeReady(&node) {
			continue
		}

		// Capacity
		if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			capacity.CPU.Add(cpu)
		}
		if memory, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			capacity.Memory.Add(memory)
		}
		if gpu, ok := node.Status.Capacity["nvidia.com/gpu"]; ok {
			capacityGPU.Add(gpu)
			hasGPU = true
		}

		// Allocatable
		if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			allocatable.CPU.Add(cpu)
		}
		if memory, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			allocatable.Memory.Add(memory)
		}
		if gpu, ok := node.Status.Allocatable["nvidia.com/gpu"]; ok {
			allocatableGPU.Add(gpu)
		}
	}

	if hasGPU {
		capacity.GPU = &capacityGPU
		allocatable.GPU = &allocatableGPU
	}

	// Calculate allocated resources from all pods
	allocated, err := c.calculateAllocatedResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate allocated resources: %w", err)
	}

	// Calculate available = allocatable - allocated
	available := &rearv1alpha1.ResourceQuantities{
		CPU:    allocatable.CPU.DeepCopy(),
		Memory: allocatable.Memory.DeepCopy(),
	}
	available.CPU.Sub(allocated.CPU)
	available.Memory.Sub(allocated.Memory)

	if hasGPU && allocated.GPU != nil {
		availableGPU := allocatable.GPU.DeepCopy()
		availableGPU.Sub(*allocated.GPU)
		available.GPU = &availableGPU
	} else if hasGPU {
		gpuCopy := allocatable.GPU.DeepCopy()
		available.GPU = &gpuCopy
	}

	return &rearv1alpha1.ResourceMetrics{
		Capacity:    *capacity,
		Allocatable: *allocatable,
		Allocated:   *allocated,
		Available:   *available,
		// Used: nil, // Can be populated with metrics-server data if available
	}, nil
}

// calculateAllocatedResources sums up all resource requests from running pods
func (c *Collector) calculateAllocatedResources(ctx context.Context) (*rearv1alpha1.ResourceQuantities, error) {
	podList := &corev1.PodList{}
	if err := c.Client.List(ctx, podList); err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	allocated := &rearv1alpha1.ResourceQuantities{
		CPU:    *resource.NewQuantity(0, resource.DecimalSI),
		Memory: *resource.NewQuantity(0, resource.BinarySI),
	}

	var allocatedGPU resource.Quantity
	hasGPU := false

	for _, pod := range podList.Items {
		// Only count pods that are running or pending
		if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
			continue
		}

		for _, container := range pod.Spec.Containers {
			// Sum CPU requests
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				allocated.CPU.Add(cpu)
			}
			// Sum Memory requests
			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				allocated.Memory.Add(memory)
			}
			// Sum GPU requests
			if gpu, ok := container.Resources.Requests["nvidia.com/gpu"]; ok {
				allocatedGPU.Add(gpu)
				hasGPU = true
			}
		}
	}

	if hasGPU {
		allocated.GPU = &allocatedGPU
	}

	return allocated, nil
}

// isNodeReady checks if a node is in Ready condition
func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// GetClusterID generates or retrieves a cluster identifier
func (c *Collector) GetClusterID(ctx context.Context) (string, error) {
	ns := &corev1.Namespace{}
	if err := c.Client.Get(ctx, client.ObjectKey{Name: "kube-system"}, ns); err != nil {
		return "", fmt.Errorf("failed to get kube-system namespace: %w", err)
	}
	return string(ns.UID), nil
}
