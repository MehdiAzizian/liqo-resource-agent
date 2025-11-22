package metrics_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/mehdiazizian/liqo-resource-agent/internal/metrics"
	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
)

func TestCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Collector Suite")
}

var _ = Describe("Metrics Collector", func() {
	var (
		ctx       context.Context
		collector *metrics.Collector
		scheme    *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = rearv1alpha1.AddToScheme(scheme)
	})

	Context("CollectClusterResources", func() {
		It("should correctly sum resources from multiple nodes", func() {
			// Create fake nodes
			node1 := createNode("node1", "4", "8Gi", true)
			node2 := createNode("node2", "4", "8Gi", true)

			// Create fake pods
			pod1 := createPod("pod1", "default", "1", "2Gi", corev1.PodRunning)

			// Create fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&node1, &node2, &pod1).
				Build()

			collector = &metrics.Collector{Client: fakeClient}

			// Collect metrics
			resourceMetrics, err := collector.CollectClusterResources(ctx)

			Expect(err).ToNot(HaveOccurred())
			Expect(resourceMetrics).ToNot(BeNil())

			// Check capacity (4+4=8 CPU)
			Expect(resourceMetrics.Capacity.CPU.String()).To(Equal("8"))
			Expect(resourceMetrics.Capacity.Memory.String()).To(Equal("16Gi"))

			// Check allocated (1 CPU from pod1)
			Expect(resourceMetrics.Allocated.CPU.String()).To(Equal("1"))
			Expect(resourceMetrics.Allocated.Memory.String()).To(Equal("2Gi"))

			// Check available (8-1=7 CPU)
			Expect(resourceMetrics.Available.CPU.String()).To(Equal("7"))
		})

		It("should ignore non-ready nodes", func() {
			node1 := createNode("node1", "4", "8Gi", true)
			node2 := createNode("node2", "4", "8Gi", false) // Not ready

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&node1, &node2).
				Build()

			collector = &metrics.Collector{Client: fakeClient}
			resourceMetrics, err := collector.CollectClusterResources(ctx)

			Expect(err).ToNot(HaveOccurred())
			// Should only count node1
			Expect(resourceMetrics.Capacity.CPU.String()).To(Equal("4"))
		})

		It("should skip terminated pods", func() {
			node1 := createNode("node1", "4", "8Gi", true)
			pod1 := createPod("pod1", "default", "1", "2Gi", corev1.PodRunning)
			pod2 := createPod("pod2", "default", "2", "4Gi", corev1.PodSucceeded)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&node1, &pod1, &pod2).
				Build()

			collector = &metrics.Collector{Client: fakeClient}
			resourceMetrics, err := collector.CollectClusterResources(ctx)

			Expect(err).ToNot(HaveOccurred())
			// Should only count pod1
			Expect(resourceMetrics.Allocated.CPU.String()).To(Equal("1"))
		})

		It("should return error when no nodes exist", func() {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			collector = &metrics.Collector{Client: fakeClient}
			_, err := collector.CollectClusterResources(ctx)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no nodes found"))
		})

		It("should handle pods with multiple containers", func() {
			node1 := createNode("node1", "10", "20Gi", true)
			pod1 := createPodWithContainers("pod1", "default", []containerResources{
				{cpu: "1", memory: "2Gi"},
				{cpu: "2", memory: "4Gi"},
			}, corev1.PodRunning)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&node1, &pod1).
				Build()

			collector = &metrics.Collector{Client: fakeClient}
			resourceMetrics, err := collector.CollectClusterResources(ctx)

			Expect(err).ToNot(HaveOccurred())
			// Should sum both containers: 1+2=3 CPU
			Expect(resourceMetrics.Allocated.CPU.String()).To(Equal("3"))
			Expect(resourceMetrics.Allocated.Memory.String()).To(Equal("6Gi"))
		})
	})

	Context("GetClusterID", func() {
		It("should return kube-system namespace UID", func() {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
					UID:  "test-cluster-123",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(ns).
				Build()

			collector = &metrics.Collector{Client: fakeClient}
			clusterID, err := collector.GetClusterID(ctx)

			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).To(Equal("test-cluster-123"))
		})
	})
})

// Helper functions

func createNode(name, cpu, memory string, ready bool) corev1.Node {
	conditions := []corev1.NodeCondition{}
	if ready {
		conditions = append(conditions, corev1.NodeCondition{
			Type:   corev1.NodeReady,
			Status: corev1.ConditionTrue,
		})
	} else {
		conditions = append(conditions, corev1.NodeCondition{
			Type:   corev1.NodeReady,
			Status: corev1.ConditionFalse,
		})
	}

	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(memory),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(memory),
			},
			Conditions: conditions,
		},
	}
}

func createPod(name, namespace, cpu, memory string, phase corev1.PodPhase) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(cpu),
							corev1.ResourceMemory: resource.MustParse(memory),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: phase,
		},
	}
}

type containerResources struct {
	cpu    string
	memory string
}

func createPodWithContainers(name, namespace string, containers []containerResources, phase corev1.PodPhase) corev1.Pod {
	podContainers := make([]corev1.Container, len(containers))
	for i, c := range containers {
		podContainers[i] = corev1.Container{
			Name: fmt.Sprintf("container%d", i+1),
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(c.cpu),
					corev1.ResourceMemory: resource.MustParse(c.memory),
				},
			},
		}
	}

	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: podContainers,
		},
		Status: corev1.PodStatus{
			Phase: phase,
		},
	}
}
