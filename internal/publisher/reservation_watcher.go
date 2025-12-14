package publisher

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReservationWatcher watches for reservations on the broker
type ReservationWatcher struct {
	Client    dynamic.Interface
	ClusterID string
	Enabled   bool
}

// NewReservationWatcher creates a new reservation watcher
func NewReservationWatcher(brokerClient *BrokerClient) *ReservationWatcher {
	if brokerClient == nil {
		return &ReservationWatcher{Enabled: false}
	}
	return &ReservationWatcher{
		Client:    brokerClient.Client,
		ClusterID: brokerClient.ClusterID,
		Enabled:   brokerClient.Enabled,
	}
}

// Start starts watching for reservations
func (w *ReservationWatcher) Start(ctx context.Context) error {
	if !w.Enabled {
		return nil
	}

	logger := log.FromContext(ctx).WithName("reservation-watcher")
	logger.Info("Starting reservation watcher", "clusterID", w.ClusterID)

	gvr := schema.GroupVersionResource{
		Group:    "broker.fluidos.eu",
		Version:  "v1alpha1",
		Resource: "reservations",
	}

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Watch reservations
			watchInterface, err := w.Client.Resource(gvr).Namespace("default").Watch(ctx, metav1.ListOptions{
				// Watch everything, we filter in loop
			})
			if err != nil {
				logger.Error(err, "Failed to start watch")
				time.Sleep(5 * time.Second)
				continue
			}

			logger.Info("Watching for reservations...")
			
			for event := range watchInterface.ResultChan() {
                // Check if it's an Unstructured object
                unstructuredObj, ok := event.Object.(*unstructured.Unstructured)
                if !ok {
                     continue 
                }
                
                spec, found, _ := unstructured.NestedMap(unstructuredObj.Object, "spec")
                if !found { continue }
                
                requesterID, found, _ := unstructured.NestedString(spec, "requesterID")
                if !found || requesterID != w.ClusterID {
                    continue
                }
                
                status, found, _ := unstructured.NestedMap(unstructuredObj.Object, "status")
                if !found { continue }
                
                phase, found, _ := unstructured.NestedString(status, "phase")
                
                // If phase is Reserved, notify!
                if phase == "Reserved" {
                    targetCluster, _, _ := unstructured.NestedString(spec, "targetClusterID")
                    
                    resources, _, _ := unstructured.NestedMap(spec, "requestedResources")
                    cpu, _, _ := unstructured.NestedString(resources, "cpu")
                    memory, _, _ := unstructured.NestedString(resources, "memory")
                    
                    logger.Info("!!! RESERVATION FULFILLED !!!", 
                        "requester", requesterID,
                        "targetCluster", targetCluster,
                        "cpu", cpu, 
                        "memory", memory)
                        
                    logger.Info(fmt.Sprintf("Manager Notification: Use %s for %s CPU, %s Memory", targetCluster, cpu, memory))
                    
                    // Here is where you would trigger Liqo peering
                    // e.g., triggerLiqoPeering(targetCluster)
                }
			}
            
            // If channel closed, retry
            logger.Info("Watch channel closed, restarting...")
            time.Sleep(1 * time.Second)
		}
	}
}
