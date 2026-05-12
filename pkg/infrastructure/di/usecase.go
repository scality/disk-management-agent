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
		pdDiscoverers := c.buildPhysicalDriveDiscoverers()
		lvDiscoverers := c.buildLogicalVolumeDiscoverers()

		c.discoverPhysicalDrivesUseCase = usecase.NewDiscoverPhysicalDrives(
			c.logger,
			pdDiscoverers,
			lvDiscoverers,
			c.getDiscoveredPhysicalDiskStore(),
			c.getDiscoveredDriveCache(),
			c.nodeName,
		)
	}

	return c.discoverPhysicalDrivesUseCase
}

// buildPhysicalDriveDiscoverers assembles the physical-drive discoverer
// slice for the use case, skipping any discoverer that could not be
// constructed (e.g. because its CLI tool is unavailable on this host).
//
// Appending a typed nil pointer to an interface slice would yield a
// non-nil interface value wrapping a nil concrete pointer, which would
// defeat the nil check inside the use case. We therefore append only
// concrete pointers that are non-nil.
func (c *Container) buildPhysicalDriveDiscoverers() []service.PhysicalDriveDiscoverer {
	var discoverers []service.PhysicalDriveDiscoverer

	if d := c.getMegaRAIDPerccliDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("MegaRAID perccli physical-drive discoverer disabled")
	}

	if d := c.getMegaRAIDStorcliDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("MegaRAID storcli physical-drive discoverer disabled")
	}

	if d := c.getSmartArrayDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("SmartArray physical-drive discoverer disabled")
	}

	return discoverers
}

// buildLogicalVolumeDiscoverers mirrors buildPhysicalDriveDiscoverers
// for the logical-volume discoverer slice. See that function for the
// rationale behind the explicit nil-check.
func (c *Container) buildLogicalVolumeDiscoverers() []service.LogicalVolumeDiscoverer {
	var discoverers []service.LogicalVolumeDiscoverer

	if d := c.getMegaRAIDPerccliLVDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("MegaRAID perccli logical-volume discoverer disabled")
	}

	if d := c.getMegaRAIDStorcliLVDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("MegaRAID storcli logical-volume discoverer disabled")
	}

	if d := c.getSmartArrayLVDiscoverer(); d != nil {
		discoverers = append(discoverers, d)
	} else {
		c.logger.Info("SmartArray logical-volume discoverer disabled")
	}

	return discoverers
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
