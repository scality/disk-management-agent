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
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metalk8sv1alpha1 "platform-disk-management-agent/api/v1alpha1"
	"platform-disk-management-agent/pkg/domain"
	"platform-disk-management-agent/pkg/service"
)

// --- Hand-written mocks ---

type mockDiscoverer struct {
	drives []*domain.DiscoveredPhysicalDrive
	err    error
}

var _ service.PhysicalDriveDiscoverer = &mockDiscoverer{}

func (m *mockDiscoverer) DiscoverPhysicalDrives() ([]*domain.DiscoveredPhysicalDrive, error) {
	return m.drives, m.err
}

type mockStore struct {
	getCalls    []string
	createCalls []*metalk8sv1alpha1.DiscoveredPhysicalDisk

	getResults    map[string]*metalk8sv1alpha1.DiscoveredPhysicalDisk
	getErr        error
	createErr     error
}

var _ service.DiscoveredPhysicalDiskStore = &mockStore{}

func (m *mockStore) Get(_ context.Context, name string) (*metalk8sv1alpha1.DiscoveredPhysicalDisk, error) {
	m.getCalls = append(m.getCalls, name)

	if m.getErr != nil {
		return nil, m.getErr
	}

	if m.getResults != nil {
		return m.getResults[name], nil
	}

	return nil, nil
}

func (m *mockStore) Create(_ context.Context, disk *metalk8sv1alpha1.DiscoveredPhysicalDisk) error {
	m.createCalls = append(m.createCalls, disk)
	return m.createErr
}

// --- Test helpers ---

func newTestHDD(ctrlType string, ctrlID int, slotID string) *domain.DiscoveredPhysicalDrive {
	slot, _ := physicaldrive.ParseSlot(slotID)

	return &domain.DiscoveredPhysicalDrive{
		ControllerType: ctrlType,
		ControllerID:   ctrlID,
		PhysicalDrive: &physicaldrive.PhysicalDrive{
			Metadata: &physicaldrive.Metadata{
				CtrlMetadata: &raidcontroller.Metadata{ID: ctrlID},
				ID:           slotID,
			},
			Slot:   slot,
			Type:   physicaldrive.DiskTypeHDD,
			Status: physicaldrive.PDStatusUsed,
			Size:   4000787030016,
			Vendor: "Seagate",
			Model:  "ST4000NM0033",
			Serial: "Z1Z2Z3Z4",
		},
	}
}

func newTestSSD(ctrlType string, ctrlID int, slotID string) *domain.DiscoveredPhysicalDrive {
	slot, _ := physicaldrive.ParseSlot(slotID)

	return &domain.DiscoveredPhysicalDrive{
		ControllerType: ctrlType,
		ControllerID:   ctrlID,
		PhysicalDrive: &physicaldrive.PhysicalDrive{
			Metadata: &physicaldrive.Metadata{
				CtrlMetadata: &raidcontroller.Metadata{ID: ctrlID},
				ID:           slotID,
			},
			Slot: slot,
			Type: physicaldrive.DiskTypeSSD,
			Size: 480103981056,
		},
	}
}

// --- Execute tests ---

func TestExecute_NoDiscoverers(t *testing.T) {
	store := &mockStore{}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), nil, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Empty(t, store.getCalls)
	assert.Empty(t, store.createCalls)
}

func TestExecute_SSDDrivesOnly(t *testing.T) {
	discoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{
			newTestSSD("MegaRAID", 0, "0:1:0"),
			newTestSSD("MegaRAID", 0, "0:1:1"),
		},
	}
	store := &mockStore{}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{discoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Empty(t, store.getCalls, "Store should not be called for SSD drives")
	assert.Empty(t, store.createCalls)
}

func TestExecute_MixOfHDDAndSSD(t *testing.T) {
	discoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{
			newTestHDD("MegaRAID", 0, "0:1:2"),
			newTestSSD("MegaRAID", 0, "0:1:3"),
		},
	}
	store := &mockStore{}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{discoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Len(t, store.getCalls, 1)
	assert.Len(t, store.createCalls, 1)
	assert.Equal(t, "node-1-megaraid-0-012", store.createCalls[0].Name)
}

func TestExecute_ExistingCR(t *testing.T) {
	hdd := newTestHDD("MegaRAID", 0, "0:1:2")
	crName := "node-1-megaraid-0-012"

	discoverer := &mockDiscoverer{drives: []*domain.DiscoveredPhysicalDrive{hdd}}
	store := &mockStore{
		getResults: map[string]*metalk8sv1alpha1.DiscoveredPhysicalDisk{
			crName: {}, // non-nil means it exists
		},
	}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{discoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{crName}, existing)
	assert.Empty(t, store.createCalls, "Should not create when CR already exists")
}

func TestExecute_DiscovererError(t *testing.T) {
	failingDiscoverer := &mockDiscoverer{err: fmt.Errorf("storcli not found")}
	store := &mockStore{}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{failingDiscoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Empty(t, store.getCalls)
}

func TestExecute_StoreGetError(t *testing.T) {
	discoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{newTestHDD("MegaRAID", 0, "0:1:2")},
	}
	store := &mockStore{getErr: fmt.Errorf("API server unavailable")}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{discoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Empty(t, store.createCalls, "Should not attempt Create when Get failed")
}

func TestExecute_StoreCreateError(t *testing.T) {
	discoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{newTestHDD("MegaRAID", 0, "0:1:2")},
	}
	store := &mockStore{createErr: fmt.Errorf("webhook rejected")}
	uc := NewDiscoverPhysicalDrives(logr.Discard(), []service.PhysicalDriveDiscoverer{discoverer}, store, "node-1")

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Len(t, store.createCalls, 1, "Create should have been attempted")
}

func TestExecute_MultipleDiscoverers(t *testing.T) {
	megaraidDiscoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{newTestHDD("MegaRAID", 0, "0:1:2")},
	}
	smartArrayDiscoverer := &mockDiscoverer{
		drives: []*domain.DiscoveredPhysicalDrive{newTestHDD("SmartArray", 1, "1I:1:4")},
	}
	store := &mockStore{}
	uc := NewDiscoverPhysicalDrives(
		logr.Discard(),
		[]service.PhysicalDriveDiscoverer{megaraidDiscoverer, smartArrayDiscoverer},
		store,
		"node-1",
	)

	existing, err := uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Empty(t, existing)
	assert.Len(t, store.createCalls, 2)

	createdNames := []string{store.createCalls[0].Name, store.createCalls[1].Name}
	assert.Contains(t, createdNames, "node-1-megaraid-0-012")
	assert.Contains(t, createdNames, "node-1-smartarray-1-1i14")
}

// --- mapPDStatus tests ---

func TestMapPDStatus(t *testing.T) {
	tests := []struct {
		status   physicaldrive.PDStatus
		expected string
	}{
		{physicaldrive.PDStatusUsed, "Used"},
		{physicaldrive.PDStatusUnassignedGood, "Available"},
		{physicaldrive.PDStatusFailed, "Failed"},
		{physicaldrive.PDStatusUnassignedBad, "Failed"},
		{physicaldrive.PDStatusUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapPDStatus(tt.status))
		})
	}
}

// --- buildCR tests ---

func TestBuildCR(t *testing.T) {
	drive := newTestHDD("MegaRAID", 0, "0:1:2")
	drive.WWN = "5000C50012345678"
	drive.JBOD = true
	drive.DevicePath = "/dev/sda"
	drive.PermanentPath = "/dev/disk/by-id/wwn-0x5000c50012345678"

	cr := buildCR("node-1-megaraid-0-012", "node-1", drive)

	assert.Equal(t, "node-1-megaraid-0-012", cr.Name)

	assert.Equal(t, "node-1", cr.Spec.NodeName)
	assert.Equal(t, "MegaRAID", cr.Spec.Controller.Type)
	assert.Equal(t, 0, cr.Spec.Controller.ID)
	assert.Equal(t, "0:1:2", cr.Spec.ID)
	assert.Equal(t, "0", cr.Spec.Slot.Port)
	assert.Equal(t, "1", cr.Spec.Slot.Enclosure)
	assert.Equal(t, "2", cr.Spec.Slot.Bay)
	assert.Equal(t, "Seagate", cr.Spec.Vendor)
	assert.Equal(t, "ST4000NM0033", cr.Spec.Model)
	assert.Equal(t, "Z1Z2Z3Z4", cr.Spec.Serial)
	assert.Equal(t, "5000C50012345678", cr.Spec.WWN)
	assert.Equal(t, int64(4000787030016), cr.Spec.Size)
	assert.Equal(t, "HDD", cr.Spec.Type)

	assert.True(t, cr.Status.JBOD)
	assert.Equal(t, "Used", cr.Status.Status)
	assert.Equal(t, "/dev/sda", cr.Status.DevicePath)
	assert.Equal(t, "/dev/disk/by-id/wwn-0x5000c50012345678", cr.Status.PermanentPath)
}

func TestBuildCR_NilSlot(t *testing.T) {
	drive := newTestHDD("MegaRAID", 0, "5")
	drive.Slot = nil

	cr := buildCR("node-1-megaraid-0-5", "node-1", drive)

	assert.Equal(t, "", cr.Spec.Slot.Port)
	assert.Equal(t, "", cr.Spec.Slot.Enclosure)
	assert.Equal(t, "", cr.Spec.Slot.Bay)
}
