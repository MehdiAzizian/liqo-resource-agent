package metrics

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
)

func buildScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("unable to add client-go scheme: %v", err)
	}
	if err := rearv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("unable to add rear scheme: %v", err)
	}
	return scheme
}

func TestCollectorAggregatesResourcesIncludingInitContainersAndOverhead(t *testing.T) {
	scheme := buildScheme(t)
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "workload",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
			InitContainers: []corev1.Container{
				{
					Name: "init-heavy",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
			Overhead: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(node, pod).
		Build()

	collector := &Collector{Client: fakeClient}
	ctx := context.Background()

	metrics, err := collector.CollectClusterResources(ctx)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	expectedAllocatedCPU := resource.MustParse("2250m") // max(0.5,2) + 0.25 = 2.25
	if metrics.Allocated.CPU.Cmp(expectedAllocatedCPU) != 0 {
		t.Fatalf("expected allocated CPU %s, got %s", expectedAllocatedCPU.String(), metrics.Allocated.CPU.String())
	}

	expectedAvailableCPU := resource.MustParse("1.75")
	if metrics.Available.CPU.Cmp(expectedAvailableCPU) != 0 {
		t.Fatalf("expected available CPU %s, got %s", expectedAvailableCPU.String(), metrics.Available.CPU.String())
	}

	expectedAllocatedMem := resource.MustParse("1152Mi")
	if metrics.Allocated.Memory.Cmp(expectedAllocatedMem) != 0 {
		t.Fatalf("expected allocated memory %s, got %s", expectedAllocatedMem.String(), metrics.Allocated.Memory.String())
	}

	expectedAvailableMem := resource.MustParse("7040Mi")
	if metrics.Available.Memory.Cmp(expectedAvailableMem) != 0 {
		t.Fatalf("expected available memory %s, got %s", expectedAvailableMem.String(), metrics.Available.Memory.String())
	}
}

func TestCollectorGetClusterIDOverride(t *testing.T) {
	collector := &Collector{
		ClusterIDOverride: "rome-cluster",
	}
	id, err := collector.GetClusterID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "rome-cluster" {
		t.Fatalf("expected override rome-cluster, got %s", id)
	}
}
