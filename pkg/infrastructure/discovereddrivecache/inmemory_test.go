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
	"testing"

	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"disk-management-agent/pkg/domain"
)

func newTestDrive(vendor string) *domain.DiscoveredPhysicalDrive {
	return &domain.DiscoveredPhysicalDrive{
		ControllerType: "MegaRAID",
		ControllerID:   0,
		PhysicalDrive: &physicaldrive.PhysicalDrive{
			Vendor: vendor,
		},
	}
}

func TestNewInMemory_NotReady(t *testing.T) {
	cache := NewInMemory()

	assert.False(t, cache.IsReady())
}

func TestLoad_BeforeReplace(t *testing.T) {
	cache := NewInMemory()

	drive, found := cache.Load("some-cr")

	assert.False(t, found)
	assert.Nil(t, drive)
}

func TestReplace_MakesCacheReady(t *testing.T) {
	cache := NewInMemory()

	cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{})

	assert.True(t, cache.IsReady())
}

func TestReplace_ThenLoad(t *testing.T) {
	cache := NewInMemory()
	drive := newTestDrive("Seagate")

	cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{
		"node-1-megaraid-0-012": drive,
	})

	got, found := cache.Load("node-1-megaraid-0-012")

	require.True(t, found)
	assert.Equal(t, "Seagate", got.Vendor)
}

func TestLoad_MissingKey(t *testing.T) {
	cache := NewInMemory()
	cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{
		"node-1-megaraid-0-012": newTestDrive("Seagate"),
	})

	_, found := cache.Load("does-not-exist")

	assert.False(t, found)
}

func TestReplace_OverwritesPreviousEntries(t *testing.T) {
	cache := NewInMemory()

	cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{
		"cr-a": newTestDrive("Seagate"),
		"cr-b": newTestDrive("WD"),
	})

	cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{
		"cr-c": newTestDrive("Toshiba"),
	})

	_, foundA := cache.Load("cr-a")
	assert.False(t, foundA, "old entry should be gone after Replace")

	got, foundC := cache.Load("cr-c")
	require.True(t, foundC)
	assert.Equal(t, "Toshiba", got.Vendor)
}
