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

package discoveredphysicaldiskstore

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metalk8sv1alpha1 "platform-disk-management-agent/api/v1alpha1"
	"platform-disk-management-agent/pkg/service"
)

// Kubernetes implements DiscoveredPhysicalDiskStore using the controller-runtime
// client to interact with the Kubernetes API server.
type Kubernetes struct {
	client client.Client
}

var _ service.DiscoveredPhysicalDiskStore = &Kubernetes{}

func NewKubernetes(c client.Client) *Kubernetes {
	return &Kubernetes{client: c}
}

// Get retrieves a DiscoveredPhysicalDisk by namespace and name.
// Returns nil and no error when the resource does not exist.
func (s *Kubernetes) Get(ctx context.Context, namespace, name string) (*metalk8sv1alpha1.DiscoveredPhysicalDisk, error) {
	disk := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}

	err := s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, disk)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return disk, nil
}

// Create persists a new DiscoveredPhysicalDisk resource.
func (s *Kubernetes) Create(ctx context.Context, disk *metalk8sv1alpha1.DiscoveredPhysicalDisk) error {
	return s.client.Create(ctx, disk)
}
