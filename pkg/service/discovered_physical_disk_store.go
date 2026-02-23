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

package service

import (
	"context"

	metalk8sv1alpha1 "platform-disk-management-agent/api/v1alpha1"
)

// DiscoveredPhysicalDiskStore abstracts Kubernetes CRUD operations on
// DiscoveredPhysicalDisk custom resources.
type DiscoveredPhysicalDiskStore interface {
	// Get retrieves a DiscoveredPhysicalDisk by name.
	// Returns nil and no error when the resource does not exist.
	Get(ctx context.Context, name string) (*metalk8sv1alpha1.DiscoveredPhysicalDisk, error)

	// Create persists a new DiscoveredPhysicalDisk resource.
	Create(ctx context.Context, disk *metalk8sv1alpha1.DiscoveredPhysicalDisk) error
}
