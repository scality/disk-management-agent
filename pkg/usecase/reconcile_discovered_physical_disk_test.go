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
	"testing"

	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

// --- Mock CacheReader ---

type mockCacheReader struct {
	ready  bool
	drives map[string]*domain.DiscoveredPhysicalDrive
}

var _ service.DiscoveredDriveCacheReader = &mockCacheReader{}

func (m *mockCacheReader) Load(crName string) (*domain.DiscoveredPhysicalDrive, bool) {
	d, ok := m.drives[crName]
	return d, ok
}

func (m *mockCacheReader) IsReady() bool {
	return m.ready
}

// --- Tests ---

func TestReconcile_CacheNotReady(t *testing.T) {
	uc := NewReconcileDiscoveredPhysicalDisk(&mockCacheReader{ready: false})

	result := uc.Execute("any-cr-name")

	assert.False(t, result.CacheReady)
	assert.False(t, result.Available)
	assert.Nil(t, result.Drive)
}

func TestReconcile_DriveNotFound(t *testing.T) {
	uc := NewReconcileDiscoveredPhysicalDisk(&mockCacheReader{
		ready:  true,
		drives: map[string]*domain.DiscoveredPhysicalDrive{},
	})

	result := uc.Execute("missing-cr")

	assert.True(t, result.CacheReady)
	assert.False(t, result.Available)
	assert.Nil(t, result.Drive)
}

func TestReconcile_DriveFound(t *testing.T) {
	drive := &domain.DiscoveredPhysicalDrive{
		ControllerType: "MegaRAID",
		ControllerID:   0,
		PhysicalDrive: &physicaldrive.PhysicalDrive{
			Metadata: &physicaldrive.Metadata{
				CtrlMetadata: &raidcontroller.Metadata{ID: 0},
				ID:           "0:1:2",
			},
			Vendor: "Seagate",
			Model:  "ST4000NM0033",
			Serial: "Z1Z2Z3Z4",
			Size:   4000787030016,
			Type:   physicaldrive.DiskTypeHDD,
			Status: physicaldrive.PDStatusUsed,
			JBOD:   true,
		},
	}

	uc := NewReconcileDiscoveredPhysicalDisk(&mockCacheReader{
		ready: true,
		drives: map[string]*domain.DiscoveredPhysicalDrive{
			"node-1-megaraid-0-012": drive,
		},
	})

	result := uc.Execute("node-1-megaraid-0-012")

	require.True(t, result.CacheReady)
	require.True(t, result.Available)
	assert.Equal(t, "Seagate", result.Drive.Vendor)
	assert.Equal(t, "ST4000NM0033", result.Drive.Model)
	assert.True(t, result.Drive.JBOD)
}
