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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
)

// nolint:unused
var discoveredphysicaldisklog = logf.Log.WithName("discoveredphysicaldisk-webhook")

// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
// +kubebuilder:object:generate=false

// DiscoveredPhysicalDiskCustomValidator validates DiscoveredPhysicalDisk CRs.
// Only the operator's service account is allowed to create or update them.
type DiscoveredPhysicalDiskCustomValidator struct {
	// AllowedServiceAccount is the full service account username
	// (e.g. "system:serviceaccount:<namespace>:<name>") that is permitted
	// to create and update DiscoveredPhysicalDisk resources.
	AllowedServiceAccount string
}

var _ admission.CustomValidator = &DiscoveredPhysicalDiskCustomValidator{}

// +kubebuilder:webhook:path=/validate-metalk8s-scality-com-v1alpha1-discoveredphysicaldisk,mutating=false,failurePolicy=fail,sideEffects=None,groups=metalk8s.scality.com,resources=discoveredphysicaldisks,verbs=create;update,versions=v1alpha1,name=vdiscoveredphysicaldisk-v1alpha1.kb.io,admissionReviewVersions=v1

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type DiscoveredPhysicalDisk.
func (v *DiscoveredPhysicalDiskCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	disk, ok := obj.(*metalk8sv1alpha1.DiscoveredPhysicalDisk)
	if !ok {
		return nil, fmt.Errorf("expected a DiscoveredPhysicalDisk object but got %T", obj)
	}
	discoveredphysicaldisklog.Info("Validation for DiscoveredPhysicalDisk upon creation", "name", disk.GetName())
	return v.validateServiceAccount(ctx)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type DiscoveredPhysicalDisk.
func (v *DiscoveredPhysicalDiskCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	disk, ok := newObj.(*metalk8sv1alpha1.DiscoveredPhysicalDisk)
	if !ok {
		return nil, fmt.Errorf("expected a DiscoveredPhysicalDisk object but got %T", newObj)
	}
	discoveredphysicaldisklog.Info("Validation for DiscoveredPhysicalDisk upon update", "name", disk.GetName())
	return v.validateServiceAccount(ctx)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type DiscoveredPhysicalDisk.
// Delete is not restricted — any user with appropriate RBAC permissions can delete these resources.
// Note: this method is required to satisfy the interface but will not be called,
// because the webhook marker above only registers for create and update verbs.
func (v *DiscoveredPhysicalDiskCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	disk, ok := obj.(*metalk8sv1alpha1.DiscoveredPhysicalDisk)
	if !ok {
		return nil, fmt.Errorf("expected a DiscoveredPhysicalDisk object but got %T", obj)
	}
	discoveredphysicaldisklog.Info("Validation for DiscoveredPhysicalDisk upon deletion", "name", disk.GetName())
	return nil, nil
}

// validateServiceAccount extracts the requesting user from the admission request
// and ensures it matches the allowed operator service account.
func (v *DiscoveredPhysicalDiskCustomValidator) validateServiceAccount(ctx context.Context) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission request in context: %w", err)
	}

	if req.UserInfo.Username != v.AllowedServiceAccount {
		return nil, fmt.Errorf(
			"only the disk management agent service account (%s) can create or update DiscoveredPhysicalDisk resources, got: %s",
			v.AllowedServiceAccount,
			req.UserInfo.Username,
		)
	}

	return nil, nil
}

// SetupDiscoveredPhysicalDiskWebhookWithManager registers the validating webhook for DiscoveredPhysicalDisk with the manager.
func SetupDiscoveredPhysicalDiskWebhookWithManager(mgr ctrl.Manager, validator *DiscoveredPhysicalDiskCustomValidator) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&metalk8sv1alpha1.DiscoveredPhysicalDisk{}).
		WithValidator(validator).
		Complete()
}
