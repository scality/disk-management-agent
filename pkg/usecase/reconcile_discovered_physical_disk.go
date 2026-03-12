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

package usecase

import (
	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

// PhysicalDriveReconcileResult carries the outcome of a single
// DiscoveredPhysicalDisk reconciliation.
type PhysicalDriveReconcileResult struct {
	// CacheReady is false when the drive cache has not been populated yet.
	CacheReady bool
	// Available indicates whether the physical drive was found in the
	// latest discovery. Only meaningful when CacheReady is true.
	Available bool
	// Drive is the discovered drive data. Nil when the drive is not
	// available or the cache is not ready.
	Drive *domain.DiscoveredPhysicalDrive
}

// ReconcileDiscoveredPhysicalDisk looks up the latest discovered state
// for a given CR name and returns the result for the controller to map
// onto the CR status.
type ReconcileDiscoveredPhysicalDisk struct {
	cacheReader service.DiscoveredDriveCacheReader
}

// NewReconcileDiscoveredPhysicalDisk creates a new reconcile use case.
func NewReconcileDiscoveredPhysicalDisk(
	cacheReader service.DiscoveredDriveCacheReader,
) *ReconcileDiscoveredPhysicalDisk {
	return &ReconcileDiscoveredPhysicalDisk{
		cacheReader: cacheReader,
	}
}

// Execute returns the reconcile result for the given CR name.
func (u *ReconcileDiscoveredPhysicalDisk) Execute(crName string) PhysicalDriveReconcileResult {
	if !u.cacheReader.IsReady() {
		return PhysicalDriveReconcileResult{CacheReady: false}
	}

	drive, found := u.cacheReader.Load(crName)
	if !found {
		return PhysicalDriveReconcileResult{CacheReady: true, Available: false}
	}

	return PhysicalDriveReconcileResult{
		CacheReady: true,
		Available:  true,
		Drive:      drive,
	}
}
