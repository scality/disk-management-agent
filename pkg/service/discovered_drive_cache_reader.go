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

import "disk-management-agent/pkg/domain"

// DiscoveredDriveCacheReader provides read-only access to the discovered
// drive cache. Used by the ReconcileDiscoveredPhysicalDisk use case.
type DiscoveredDriveCacheReader interface {
	// Load returns the cached drive for the given CR name.
	// The second return value is false when no entry exists.
	Load(crName string) (*domain.DiscoveredPhysicalDrive, bool)
	// IsReady reports whether the cache has been populated at least once.
	IsReady() bool
}
