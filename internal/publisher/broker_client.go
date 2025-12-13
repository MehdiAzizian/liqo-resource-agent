package publisher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
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

// PublishAdvertisement publishes or updates an advertisement in the broker with retry logic
func (b *BrokerClient) PublishAdvertisement(ctx context.Context, adv *rearv1alpha1.Advertisement) error {
	if !b.Enabled {
		return nil
	}

	// Retry with exponential backoff
	backoff := wait.Backoff{
		Steps:    4,
		Duration: 500 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}

	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := b.publishOnce(ctx, adv)
		if err == nil {
			return true, nil // Success
		}

		// Check if error is transient
		if isTransientError(err) {
			lastErr = err
			return false, nil // Retry
		}

		// Permanent error
		return false, err
	})

	if err != nil {
		if lastErr != nil {
			return fmt.Errorf("publish failed after retries: %w", lastErr)
		}
		return err
	}

	return nil
}

// publishOnce performs a single publish attempt
func (b *BrokerClient) publishOnce(ctx context.Context, adv *rearv1alpha1.Advertisement) error {
	// Define the GVR for ClusterAdvertisement
	gvr := schema.GroupVersionResource{
		Group:    "broker.fluidos.eu",
		Version:  "v1alpha1",
		Resource: "clusteradvertisements",
	}

	resourceClient := b.Client.Resource(gvr).Namespace("default")

	// Try to get existing to preserve Reserved field
	existing, err := resourceClient.Get(ctx, fmt.Sprintf("%s-adv", b.ClusterID), metav1.GetOptions{})

	var reservedResources map[string]interface{}
	if err == nil && existing != nil {
		// Preserve existing reserved resources
		if spec, ok := existing.Object["spec"].(map[string]interface{}); ok {
			if resources, ok := spec["resources"].(map[string]interface{}); ok {
				if reserved, ok := resources["reserved"].(map[string]interface{}); ok {
					reservedResources = reserved
				}
			}
		}
	}

	// Build resources spec
	resourcesSpec := map[string]interface{}{
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
	}

	// Add reserved if it existed
	if reservedResources != nil {
		resourcesSpec["reserved"] = reservedResources
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
				"resources":   resourcesSpec,
				"timestamp":   adv.Spec.Timestamp.Format("2006-01-02T15:04:05Z"),
			},
		},
	}

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

// isTransientError checks if error is temporary and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Kubernetes API errors
	if apierrors.IsTimeout(err) ||
		apierrors.IsServerTimeout(err) ||
		apierrors.IsServiceUnavailable(err) ||
		apierrors.IsTooManyRequests(err) ||
		apierrors.IsInternalError(err) {
		return true
	}

	return false
}

// ReservationEventHandler is called when a reservation event occurs
type ReservationEventHandler func(eventType string, reservation map[string]interface{})

// WatchReservationsForCluster watches the broker for reservation events where requesterID matches this cluster
// This provides instant notification when the broker makes a decision (no polling delay)
func (b *BrokerClient) WatchReservationsForCluster(ctx context.Context, handler ReservationEventHandler) {
	if !b.Enabled {
		return
	}

	// Define GVR for Reservation
	gvr := schema.GroupVersionResource{
		Group:    "broker.fluidos.eu",
		Version:  "v1alpha1",
		Resource: "reservations",
	}

	resourceClient := b.Client.Resource(gvr).Namespace("default")

	// Start watching in a goroutine
	go func() {
		for {
			// Create watch
			watcher, err := resourceClient.Watch(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to start watch: %v, retrying in 5s...\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Process watch events
			for event := range watcher.ResultChan() {
				if event.Object == nil {
					continue
				}

				// Convert to unstructured
				obj, ok := event.Object.(*unstructured.Unstructured)
				if !ok {
					continue
				}

				// Check if requesterID matches our cluster
				spec, ok := obj.Object["spec"].(map[string]interface{})
				if !ok {
					continue
				}

				requesterID, ok := spec["requesterID"].(string)
				if !ok || requesterID != b.ClusterID {
					continue // Not for us
				}

				// Call handler with event type and reservation data
				handler(string(event.Type), obj.Object)
			}

			// Watch channel closed, reconnect after delay
			fmt.Println("Watch connection closed, reconnecting in 5s...")
			time.Sleep(5 * time.Second)
		}
	}()
}
