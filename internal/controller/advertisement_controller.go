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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rearv1alpha1 "github.com/mehdiazizian/tesi2-ra/api/v1alpha1"
	"github.com/mehdiazizian/tesi2-ra/internal/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdvertisementReconciler reconciles an Advertisement object
type AdvertisementReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	MetricsCollector *metrics.Collector
}

// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rear.fluidos.eu,resources=advertisements/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AdvertisementReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Advertisement", "name", req.Name, "namespace", req.Namespace)

	// Fetch the Advertisement instance
	advertisement := &rearv1alpha1.Advertisement{}
	err := r.Get(ctx, req.NamespacedName, advertisement)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Advertisement not found, could have been deleted
			logger.Info("Advertisement not found, may have been deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Advertisement")
		return ctrl.Result{}, err
	}

	// Collect current cluster metrics
	resourceData, err := r.MetricsCollector.CollectClusterResources(ctx)
	if err != nil {
		logger.Error(err, "Failed to collect cluster resources")
		return r.updateStatus(ctx, advertisement, "Error", false, fmt.Sprintf("Failed to collect metrics: %v", err))
	}

	// Get cluster ID
	clusterID, err := r.MetricsCollector.GetClusterID(ctx)
	if err != nil {
		logger.Error(err, "Failed to get cluster ID")
		return r.updateStatus(ctx, advertisement, "Error", false, fmt.Sprintf("Failed to get cluster ID: %v", err))
	}

	// Update the Advertisement spec with collected data
	advertisement.Spec.ClusterID = clusterID
	advertisement.Spec.Resources = *resourceData
	advertisement.Spec.Timestamp = metav1.Now()

	// Update the Advertisement resource
	if err := r.Update(ctx, advertisement); err != nil {
		logger.Error(err, "Failed to update Advertisement spec")
		return ctrl.Result{}, err
	}

	logger.Info("Updated Advertisement with current metrics",
		"capacity-cpu", resourceData.Capacity.CPU.String(),
		"allocatable-cpu", resourceData.Allocatable.CPU.String(),
		"allocated-cpu", resourceData.Allocated.CPU.String(),
		"available-cpu", resourceData.Available.CPU.String(),
		"allocatable-memory", resourceData.Allocatable.Memory.String(),
		"allocated-memory", resourceData.Allocated.Memory.String())

	// Update status to indicate successful publication
	return r.updateStatus(ctx, advertisement, "Active", true, "Advertisement updated successfully")
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
		logger.Error(err, "Failed to update Advertisement status")
		return ctrl.Result{}, err
	}

	// Requeue after 30 seconds for periodic updates
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AdvertisementReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize metrics collector if not set
	if r.MetricsCollector == nil {
		r.MetricsCollector = &metrics.Collector{
			Client: r.Client,
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&rearv1alpha1.Advertisement{}).
		Named("advertisement").
		Complete(r)
}
