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

package di

import (
	"disk-management-agent/pkg/infrastructure/discovereddrivecache"
	"disk-management-agent/pkg/infrastructure/discoveredphysicaldiskstore"
	"disk-management-agent/pkg/service"
	"disk-management-agent/pkg/usecase"
)

func (c *Container) getDiscoveredPhysicalDiskStore() *discoveredphysicaldiskstore.Kubernetes {
	if c.discoveredPhysicalDiskStore == nil {
		c.discoveredPhysicalDiskStore = discoveredphysicaldiskstore.NewKubernetes(c.k8sClient)
	}

	return c.discoveredPhysicalDiskStore
}

func (c *Container) getDiscoveredDriveCache() *discovereddrivecache.InMemory {
	if c.discoveredDriveCache == nil {
		c.discoveredDriveCache = discovereddrivecache.NewInMemory()
	}

	return c.discoveredDriveCache
}

// GetDiscoverPhysicalDrivesUseCase returns the singleton use case instance.
func (c *Container) GetDiscoverPhysicalDrivesUseCase() *usecase.DiscoverPhysicalDrives {
	if c.discoverPhysicalDrivesUseCase == nil {
		discoverers := []service.PhysicalDriveDiscoverer{
			c.getMegaRAIDPerccliDiscoverer(),
			c.getMegaRAIDStorcliDiscoverer(),
			c.getSmartArrayDiscoverer(),
		}

		c.discoverPhysicalDrivesUseCase = usecase.NewDiscoverPhysicalDrives(
			c.logger,
			discoverers,
			c.getDiscoveredPhysicalDiskStore(),
			c.getDiscoveredDriveCache(),
			c.nodeName,
		)
	}

	return c.discoverPhysicalDrivesUseCase
}

// GetReconcileDiscoveredPhysicalDiskUseCase returns the singleton reconcile use case.
func (c *Container) GetReconcileDiscoveredPhysicalDiskUseCase() *usecase.ReconcileDiscoveredPhysicalDisk {
	if c.reconcileDiscoveredPhysicalDiskUC == nil {
		c.reconcileDiscoveredPhysicalDiskUC = usecase.NewReconcileDiscoveredPhysicalDisk(
			c.getDiscoveredDriveCache(),
		)
	}

	return c.reconcileDiscoveredPhysicalDiskUC
}
