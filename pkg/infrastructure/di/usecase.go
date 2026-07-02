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
	"github.com/go-logr/logr"

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

// discovererCandidate names a discoverer and the getter that constructs it.
// get returns a concrete pointer type: a nil pointer means the discoverer
// could not be built (its CLI tool is absent on this host). The bool reports
// whether the returned interface value is present — the check must run on the
// concrete pointer inside get, because appending a typed nil pointer to an
// interface slice would yield a non-nil interface wrapping it, defeating the
// nil check inside the use case.
type discovererCandidate[T any] struct {
	kind string
	get  func() (T, bool)
}

// collectDiscoverers appends every present candidate, logging the rest.
func collectDiscoverers[T any](logger logr.Logger, role string, candidates []discovererCandidate[T]) []T {
	var discoverers []T

	for _, candidate := range candidates {
		if d, ok := candidate.get(); ok {
			discoverers = append(discoverers, d)
		} else {
			logger.Info(candidate.kind + " " + role + " discoverer disabled")
		}
	}

	return discoverers
}

// present adapts a concrete-pointer getter to the (value, ok) shape
// collectDiscoverers expects. Keeping the nil check on the concrete pointer
// is what makes it meaningful: a nil *T assigned to interface I yields a
// non-nil I, so we report ok from the concrete pointer, not the interface.
func present[I any, T any](d *T) (I, bool) {
	if d != nil {
		return any(d).(I), true
	}

	var zero I

	return zero, false
}

// buildPhysicalDriveDiscoverers assembles the physical-drive discoverer
// slice for the use case, skipping any discoverer that could not be
// constructed (e.g. because its CLI tool is unavailable on this host).
func (c *Container) buildPhysicalDriveDiscoverers() []service.PhysicalDriveDiscoverer {
	type discoverer = service.PhysicalDriveDiscoverer

	return collectDiscoverers(c.logger, "physical-drive", []discovererCandidate[discoverer]{
		{"MegaRAID perccli", func() (discoverer, bool) { return present[discoverer](c.getMegaRAIDPerccliDiscoverer()) }},
		{"MegaRAID storcli", func() (discoverer, bool) { return present[discoverer](c.getMegaRAIDStorcliDiscoverer()) }},
		{"storcli2", func() (discoverer, bool) { return present[discoverer](c.getStorcli2Discoverer()) }},
		{"perccli2", func() (discoverer, bool) { return present[discoverer](c.getPerccli2Discoverer()) }},
		{"SmartArray", func() (discoverer, bool) { return present[discoverer](c.getSmartArrayDiscoverer()) }},
	})
}

// buildLogicalVolumeDiscoverers mirrors buildPhysicalDriveDiscoverers
// for the logical-volume discoverer slice.
//
//nolint:dupl // Mirrors buildPhysicalDriveDiscoverers by design.
func (c *Container) buildLogicalVolumeDiscoverers() []service.LogicalVolumeDiscoverer {
	type discoverer = service.LogicalVolumeDiscoverer

	return collectDiscoverers(c.logger, "logical-volume", []discovererCandidate[discoverer]{
		{"MegaRAID perccli", func() (discoverer, bool) { return present[discoverer](c.getMegaRAIDPerccliLVDiscoverer()) }},
		{"MegaRAID storcli", func() (discoverer, bool) { return present[discoverer](c.getMegaRAIDStorcliLVDiscoverer()) }},
		{"storcli2", func() (discoverer, bool) { return present[discoverer](c.getStorcli2LVDiscoverer()) }},
		{"perccli2", func() (discoverer, bool) { return present[discoverer](c.getPerccli2LVDiscoverer()) }},
		{"SmartArray", func() (discoverer, bool) { return present[discoverer](c.getSmartArrayLVDiscoverer()) }},
	})
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
