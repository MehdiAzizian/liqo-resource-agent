package publisher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
)

// BrokerClient publishes advertisements to the broker cluster
type BrokerClient struct {
	Client    dynamic.Interface
	ClusterID string
	Enabled   bool
}

// NewBrokerClient creates a new broker client using dynamic client
func NewBrokerClient(brokerKubeconfig, clusterID string) (*BrokerClient, error) {
	// Check if broker publishing is enabled
	if brokerKubeconfig == "" {
		return &BrokerClient{
			Enabled: false,
		}, nil
	}

	// Load kubeconfig for broker cluster
	config, err := loadBrokerConfig(brokerKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load broker kubeconfig: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &BrokerClient{
		Client:    dynamicClient,
		ClusterID: clusterID,
		Enabled:   true,
	}, nil
}

// loadBrokerConfig loads kubeconfig from file
func loadBrokerConfig(kubeconfigPath string) (*rest.Config, error) {
	// Expand ~ to home directory
	if len(kubeconfigPath) >= 2 && kubeconfigPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfigPath = filepath.Join(home, kubeconfigPath[2:])
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// PublishAdvertisement publishes or updates an advertisement in the broker
func (b *BrokerClient) PublishAdvertisement(ctx context.Context, adv *rearv1alpha1.Advertisement) error {
	if !b.Enabled {
		// Publishing disabled, skip
		return nil
	}

	// Define the GVR for ClusterAdvertisement
	gvr := schema.GroupVersionResource{
		Group:    "broker.fluidos.eu",
		Version:  "v1alpha1",
		Resource: "clusteradvertisements",
	}

	// Convert to unstructured
	clusterAdv := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "broker.fluidos.eu/v1alpha1",
			"kind":       "ClusterAdvertisement",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-adv", b.ClusterID),
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"clusterID":   adv.Spec.ClusterID,
				"clusterName": b.ClusterID,
				"resources": map[string]interface{}{
					"capacity": map[string]interface{}{
						"cpu":    adv.Spec.Resources.Capacity.CPU.String(),
						"memory": adv.Spec.Resources.Capacity.Memory.String(),
					},
					"allocatable": map[string]interface{}{
						"cpu":    adv.Spec.Resources.Allocatable.CPU.String(),
						"memory": adv.Spec.Resources.Allocatable.Memory.String(),
					},
					"allocated": map[string]interface{}{
						"cpu":    adv.Spec.Resources.Allocated.CPU.String(),
						"memory": adv.Spec.Resources.Allocated.Memory.String(),
					},
					"available": map[string]interface{}{
						"cpu":    adv.Spec.Resources.Available.CPU.String(),
						"memory": adv.Spec.Resources.Available.Memory.String(),
					},
				},
				"timestamp": adv.Spec.Timestamp.Format("2006-01-02T15:04:05Z"),
			},
		},
	}

	// Try to get existing
	resourceClient := b.Client.Resource(gvr).Namespace("default")
	existing, err := resourceClient.Get(ctx, clusterAdv.GetName(), metav1.GetOptions{})

	if err != nil {
		// Doesn't exist, create
		_, err = resourceClient.Create(ctx, clusterAdv, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create advertisement in broker: %w", err)
		}
		return nil
	}

	// Exists, update with preserved resourceVersion
	existingMeta := existing.Object["metadata"].(map[string]interface{})
	clusterAdvMeta := clusterAdv.Object["metadata"].(map[string]interface{})
	clusterAdvMeta["resourceVersion"] = existingMeta["resourceVersion"]

	_, err = resourceClient.Update(ctx, clusterAdv, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update advertisement in broker: %w", err)
	}

	return nil
}
