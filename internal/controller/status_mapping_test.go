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

package controller

import (
	"testing"

	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"
	"github.com/stretchr/testify/assert"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/usecase"
)

func TestMapDriveToStatus_Available(t *testing.T) {
	result := usecase.PhysicalDriveReconcileResult{
		CacheReady: true,
		Available:  true,
		Drive: &domain.DiscoveredPhysicalDrive{
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
				WWN:    "5000C50012345678",
				Size:   4000787030016,
				Type:   physicaldrive.DiskTypeHDD,
				Status: physicaldrive.PDStatusUsed,
				JBOD:   true,
				Reason: "part of logical volume",
			},
		},
	}

	status := &metalk8sv1alpha1.DiscoveredPhysicalDiskStatus{}
	mapDriveToStatus(status, result)

	assert.True(t, *status.Available)
	assert.Equal(t, "Seagate", *status.Vendor)
	assert.Equal(t, "ST4000NM0033", *status.Model)
	assert.Equal(t, "Z1Z2Z3Z4", *status.Serial)
	assert.Equal(t, "5000C50012345678", *status.WWN)
	assert.Equal(t, uint64(4000787030016), *status.Size)
	assert.Equal(t, "HDD", *status.Type)
	assert.True(t, *status.JBOD)
	assert.Equal(t, "Used", *status.Status)
	assert.Equal(t, "part of logical volume", *status.Reason)
}

func TestMapDriveToStatus_Unavailable(t *testing.T) {
	result := usecase.PhysicalDriveReconcileResult{
		CacheReady: true,
		Available:  false,
	}

	status := &metalk8sv1alpha1.DiscoveredPhysicalDiskStatus{}
	mapDriveToStatus(status, result)

	assert.False(t, *status.Available)
	assert.Nil(t, status.Vendor)
	assert.Nil(t, status.Model)
}
