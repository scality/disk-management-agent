/*
Copyright 2026.

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

	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
	"disk-management-agent/pkg/usecase"
)

const cacheNotReadyRequeueDelay = 30 * time.Second

// DiscoveredPhysicalDiskReconciler reconciles a DiscoveredPhysicalDisk object
type DiscoveredPhysicalDiskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// NodeName is the name of the Kubernetes node this agent is running on.
	// The manager cache is configured with a field selector so that only
	// DiscoveredPhysicalDisk resources whose spec.nodeName matches this value
	// are watched and cached.
	NodeName string
	// ReconcileUseCase provides the discovered drive state for a given CR.
	ReconcileUseCase *usecase.ReconcileDiscoveredPhysicalDisk
}

// +kubebuilder:rbac:groups=metalk8s.scality.com,resources=discoveredphysicaldisks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=metalk8s.scality.com,resources=discoveredphysicaldisks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=metalk8s.scality.com,resources=discoveredphysicaldisks/finalizers,verbs=update

// Reconcile fetches the latest discovered drive data from the cache and
// updates the CR status accordingly.
func (r *DiscoveredPhysicalDiskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling DiscoveredPhysicalDisk")

	cr := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	result := r.ReconcileUseCase.Execute(req.Name)

	if !result.CacheReady {
		logger.Info("Drive cache not ready yet, requeuing")
		return ctrl.Result{RequeueAfter: cacheNotReadyRequeueDelay}, nil
	}

	mapDriveToStatus(&cr.Status, result)

	if err := r.Status().Update(ctx, cr); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to update DiscoveredPhysicalDisk status: %w", err)
	}

	logger.Info("DiscoveredPhysicalDisk status updated")

	return ctrl.Result{}, nil
}

// mapDriveToStatus writes the reconcile result onto the CR status.
func mapDriveToStatus(status *metalk8sv1alpha1.DiscoveredPhysicalDiskStatus, result usecase.PhysicalDriveReconcileResult) {
	status.Available = ptr(result.Available)

	if !result.Available {
		return
	}

	drive := result.Drive

	status.Vendor = &drive.Vendor
	status.Model = &drive.Model
	status.Serial = &drive.Serial
	status.WWN = &drive.WWN
	status.Size = ptr(int64(drive.Size)) //nolint:gosec // raidmgmt uses uint64; overflow is not a concern for disk sizes
	status.Type = ptr(drive.Type.String())
	status.JBOD = &drive.JBOD
	status.Status = ptr(mapPDStatus(drive.Status))
	status.Reason = &drive.Reason
}

func mapPDStatus(status physicaldrive.PDStatus) string {
	switch status {
	case physicaldrive.PDStatusUsed:
		return "Used"
	case physicaldrive.PDStatusUnassignedGood:
		return "Available"
	case physicaldrive.PDStatusFailed, physicaldrive.PDStatusUnassignedBad:
		return "Failed"
	default:
		return ""
	}
}

func ptr[T any](v T) *T {
	return &v
}

// SetupWithManager sets up the controller with the Manager.
// tickerEvents is a channel of GenericEvent sent by the DiscoveryTicker;
// receiving on this channel will trigger reconciliation.
func (r *DiscoveredPhysicalDiskReconciler) SetupWithManager(mgr ctrl.Manager, tickerEvents <-chan event.GenericEvent) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&metalk8sv1alpha1.DiscoveredPhysicalDisk{}).
		WatchesRawSource(source.Channel(tickerEvents, &handler.EnqueueRequestForObject{})).
		Named("discoveredphysicaldisk").
		Complete(r)
}
