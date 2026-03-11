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
	"testing"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
)

const (
	allowedSA        = "system:serviceaccount:test-ns:test-sa"
	unauthorizedUser = "system:serviceaccount:other-ns:other-sa"
)

func newTestValidator() *DiscoveredPhysicalDiskCustomValidator {
	return &DiscoveredPhysicalDiskCustomValidator{
		AllowedServiceAccount: allowedSA,
	}
}

func newTestDisk() *metalk8sv1alpha1.DiscoveredPhysicalDisk {
	return &metalk8sv1alpha1.DiscoveredPhysicalDisk{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-disk",
			Namespace: "default",
		},
		Spec: metalk8sv1alpha1.DiscoveredPhysicalDiskSpec{
			NodeName: "node-1",
			Controller: metalk8sv1alpha1.ControllerRef{
				Type: "MegaRAID",
				ID:   0,
			},
			ID: "0:1:2",
			Slot: metalk8sv1alpha1.SlotLocation{
				Port:      "0",
				Enclosure: "1",
				Bay:       "2",
			},
			Size: 4000787030016,
			Type: "HDD",
		},
	}
}

func contextWithUser(username string) context.Context {
	req := admission.Request{}
	req.UserInfo = authenticationv1.UserInfo{
		Username: username,
	}
	return admission.NewContextWithRequest(context.Background(), req)
}

func TestValidateCreate_AllowedServiceAccount(t *testing.T) {
	v := newTestValidator()
	ctx := contextWithUser(allowedSA)

	warnings, err := v.ValidateCreate(ctx, newTestDisk())
	if err != nil {
		t.Errorf("expected no error for allowed service account, got: %v", err)
	}
	if warnings != nil {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestValidateCreate_UnauthorizedUser(t *testing.T) {
	v := newTestValidator()
	ctx := contextWithUser(unauthorizedUser)

	_, err := v.ValidateCreate(ctx, newTestDisk())
	if err == nil {
		t.Error("expected error for unauthorized user, got nil")
	}
}

func TestValidateCreate_RegularUser(t *testing.T) {
	v := newTestValidator()
	ctx := contextWithUser("user:admin")

	_, err := v.ValidateCreate(ctx, newTestDisk())
	if err == nil {
		t.Error("expected error for regular user, got nil")
	}
}

func TestValidateUpdate_AllowedServiceAccount(t *testing.T) {
	v := newTestValidator()
	ctx := contextWithUser(allowedSA)
	disk := newTestDisk()

	warnings, err := v.ValidateUpdate(ctx, disk, disk)
	if err != nil {
		t.Errorf("expected no error for allowed service account, got: %v", err)
	}
	if warnings != nil {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestValidateUpdate_UnauthorizedUser(t *testing.T) {
	v := newTestValidator()
	ctx := contextWithUser(unauthorizedUser)
	disk := newTestDisk()

	_, err := v.ValidateUpdate(ctx, disk, disk)
	if err == nil {
		t.Error("expected error for unauthorized user, got nil")
	}
}

func TestValidateDelete_AnyUser(t *testing.T) {
	v := newTestValidator()
	// ValidateDelete should always succeed regardless of the caller.
	// Note: in practice this method won't be called because the webhook
	// is only registered for create and update verbs.
	ctx := contextWithUser(unauthorizedUser)

	warnings, err := v.ValidateDelete(ctx, newTestDisk())
	if err != nil {
		t.Errorf("expected no error on delete, got: %v", err)
	}
	if warnings != nil {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestValidateCreate_NoAdmissionRequest(t *testing.T) {
	v := newTestValidator()
	ctx := context.Background()

	_, err := v.ValidateCreate(ctx, newTestDisk())
	if err == nil {
		t.Error("expected error when no admission request in context, got nil")
	}
}
