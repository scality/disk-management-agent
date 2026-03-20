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

package discovereddrivecache

import (
	"sync"

	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

// InMemory is a thread-safe in-memory cache that stores the latest set of
// discovered physical drives keyed by CR name. It implements both the
// writer and reader interfaces so the discovery and reconcile paths can
// share the same instance through their respective narrow interfaces.
type InMemory struct {
	mu     sync.RWMutex
	drives map[string]*domain.DiscoveredPhysicalDrive
	ready  bool
}

var (
	_ service.DiscoveredDriveCacheWriter = &InMemory{}
	_ service.DiscoveredDriveCacheReader = &InMemory{}
)

func NewInMemory() *InMemory {
	return &InMemory{}
}

// Replace atomically swaps the entire cache contents with the supplied map.
func (c *InMemory) Replace(drives map[string]*domain.DiscoveredPhysicalDrive) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.drives = drives
	c.ready = true
}

// Load returns the cached drive for the given CR name.
func (c *InMemory) Load(crName string) (*domain.DiscoveredPhysicalDrive, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	drive, ok := c.drives[crName]

	return drive, ok
}

// IsReady reports whether Replace has been called at least once.
func (c *InMemory) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.ready
}
