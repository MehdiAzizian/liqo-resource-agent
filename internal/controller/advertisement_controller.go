/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
	"github.com/mehdiazizian/liqo-resource-agent/internal/metrics"
	"github.com/mehdiazizian/liqo-resource-agent/internal/publisher" // ← Add this line
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdvertisementReconciler reconciles an Advertisement object
type AdvertisementReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetricsCollector *metrics.Collector
	BrokerClient     *publisher.BrokerClient
	TargetKey        types.NamespacedName
	RequeueInterval  time.Duration // Configurable requeue interval
}

// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AdvertisementReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling advertisement",
		"name", req.Name,
		"namespace", req.Namespace)

	// Fetch the Advertisement instance
	advertisement := &rearv1alpha1.Advertisement{}
	err := r.Get(ctx, req.NamespacedName, advertisement)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("advertisement not found, may have been deleted",
				"name", req.Name,
				"namespace", req.Namespace)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get advertisement",
			"name", req.Name,
			"namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	// Collect current cluster metrics
	resourceData, err := r.MetricsCollector.CollectClusterResources(ctx)
	if err != nil {
		logger.Error(err, "failed to collect cluster resources")
		return r.updateStatus(ctx, advertisement, "Error", false, fmt.Sprintf("Failed to collect metrics: %v", err))
	}

	// Get cluster ID
	clusterID, err := r.MetricsCollector.GetClusterID(ctx)
	if err != nil {
		logger.Error(err, "failed to get cluster ID")
		return r.updateStatus(ctx, advertisement, "Error", false, fmt.Sprintf("Failed to get cluster ID: %v", err))
	}

	// Update the Advertisement spec with collected data
	advertisement.Spec.ClusterID = clusterID
	advertisement.Spec.Resources = *resourceData
	advertisement.Spec.Timestamp = metav1.Now()

	// Update the Advertisement resource
	if err := r.Update(ctx, advertisement); err != nil {
		logger.Error(err, "failed to update advertisement spec",
			"name", advertisement.Name,
			"namespace", advertisement.Namespace)
		return ctrl.Result{}, err
	}

	logger.Info("updated advertisement with current metrics",
		"clusterID", clusterID,
		"capacityCPU", resourceData.Capacity.CPU.String(),
		"allocatableCPU", resourceData.Allocatable.CPU.String(),
		"allocatedCPU", resourceData.Allocated.CPU.String(),
		"availableCPU", resourceData.Available.CPU.String(),
		"allocatableMemory", resourceData.Allocatable.Memory.String(),
		"allocatedMemory", resourceData.Allocated.Memory.String(),
		"availableMemory", resourceData.Available.Memory.String())

	// Update status to indicate successful publication
	result, err := r.updateStatus(ctx, advertisement, "Active", true, "Advertisement updated successfully")

	// Publish to broker if enabled
	if r.BrokerClient != nil && r.BrokerClient.Enabled {
		if err := r.BrokerClient.PublishAdvertisement(ctx, advertisement); err != nil {
			logger.Error(err, "failed to publish to broker, will retry",
				"clusterID", clusterID)
			// Don't fail the reconciliation, just log the error
		} else {
			logger.Info("successfully published to broker",
				"clusterID", clusterID,
				"advertisementName", advertisement.Name)
		}
	}

	return result, err
}

// updateStatus updates the Advertisement status
func (r *AdvertisementReconciler) updateStatus(
	ctx context.Context,
	advertisement *rearv1alpha1.Advertisement,
	phase string,
	published bool,
	message string,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	advertisement.Status.Phase = phase
	advertisement.Status.Published = published
	advertisement.Status.Message = message
	advertisement.Status.LastUpdateTime = metav1.Now()

	if err := r.Status().Update(ctx, advertisement); err != nil {
		logger.Error(err, "failed to update advertisement status",
			"name", advertisement.Name,
			"namespace", advertisement.Namespace,
			"phase", phase)
		return ctrl.Result{}, err
	}

	// Requeue for periodic updates
	requeueInterval := r.RequeueInterval
	if requeueInterval == 0 {
		requeueInterval = 30 * time.Second // Default if not configured
	}
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AdvertisementReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize metrics collector if not set
	if r.MetricsCollector == nil {
		r.MetricsCollector = &metrics.Collector{}
	}
	r.MetricsCollector.Client = r.Client

	return ctrl.NewControllerManagedBy(mgr).
		For(&rearv1alpha1.Advertisement{}).
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(r.findAdvertisementsForNode),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findAdvertisementsForPod),
		).
		Named("advertisement").
		Complete(r)
}

// findAdvertisementsForNode triggers reconciliation when nodes change
func (r *AdvertisementReconciler) findAdvertisementsForNode(ctx context.Context, node client.Object) []reconcile.Request {
	if r.TargetKey.Name == "" {
		return nil
	}
	return []reconcile.Request{{NamespacedName: r.TargetKey}}
}

// findAdvertisementsForPod triggers reconciliation when pods change
func (r *AdvertisementReconciler) findAdvertisementsForPod(ctx context.Context, pod client.Object) []reconcile.Request {
	if r.TargetKey.Name == "" {
		return nil
	}
	return []reconcile.Request{{NamespacedName: r.TargetKey}}
}
