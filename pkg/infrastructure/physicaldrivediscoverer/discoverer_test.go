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

package physicaldrivediscoverer

import (
	"fmt"
	"testing"

	"github.com/scality/raidmgmt/pkg/domain/entities/logicalvolume"
	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"disk-management-agent/pkg/service"
)

// mockRAIDController implements ports.RAIDController for testing.
// Only Controllers and PhysicalDrives are exercised by the discoverers;
// every other method panics to surface accidental calls.
type mockRAIDController struct {
	mock.Mock
}

func (m *mockRAIDController) Controllers() ([]*raidcontroller.RAIDController, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*raidcontroller.RAIDController), args.Error(1)
}

func (m *mockRAIDController) Controller(_ *raidcontroller.Metadata) (*raidcontroller.RAIDController, error) {
	panic("unexpected call to Controller")
}

func (m *mockRAIDController) PhysicalDrives(meta *raidcontroller.Metadata) ([]*physicaldrive.PhysicalDrive, error) {
	args := m.Called(meta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*physicaldrive.PhysicalDrive), args.Error(1)
}

func (m *mockRAIDController) PhysicalDrive(_ *physicaldrive.Metadata) (*physicaldrive.PhysicalDrive, error) {
	panic("unexpected call to PhysicalDrive")
}

func (m *mockRAIDController) LogicalVolumes(_ *raidcontroller.Metadata) ([]*logicalvolume.LogicalVolume, error) {
	panic("unexpected call to LogicalVolumes")
}

func (m *mockRAIDController) LogicalVolume(_ *logicalvolume.Metadata) (*logicalvolume.LogicalVolume, error) {
	panic("unexpected call to LogicalVolume")
}

func (m *mockRAIDController) CreateLV(_ *logicalvolume.Request) (*logicalvolume.LogicalVolume, error) {
	panic("unexpected call to CreateLV")
}

func (m *mockRAIDController) DeleteLV(_ *logicalvolume.Metadata) error {
	panic("unexpected call to DeleteLV")
}

func (m *mockRAIDController) AddPDsToLV(_ *logicalvolume.Metadata, _ ...*physicaldrive.Metadata) error {
	panic("unexpected call to AddPDsToLV")
}

func (m *mockRAIDController) DeletePDsFromLV(_ *logicalvolume.Metadata, _ ...*physicaldrive.Metadata) error {
	panic("unexpected call to DeletePDsFromLV")
}

func (m *mockRAIDController) SetLVCacheOptions(_ *logicalvolume.Metadata, _ *logicalvolume.CacheOptions) error {
	panic("unexpected call to SetLVCacheOptions")
}

func (m *mockRAIDController) EnableJBOD(_ *physicaldrive.Metadata) error {
	panic("unexpected call to EnableJBOD")
}

func (m *mockRAIDController) DisableJBOD(_ *physicaldrive.Metadata) error {
	panic("unexpected call to DisableJBOD")
}

func (m *mockRAIDController) StartBlink(_ *physicaldrive.Metadata) error {
	panic("unexpected call to StartBlink")
}

func (m *mockRAIDController) StopBlink(_ *physicaldrive.Metadata) error {
	panic("unexpected call to StopBlink")
}

func newController(id int) *raidcontroller.RAIDController {
	return &raidcontroller.RAIDController{
		Metadata: &raidcontroller.Metadata{ID: id},
	}
}

func newDrive(id string, vendor string) *physicaldrive.PhysicalDrive {
	return &physicaldrive.PhysicalDrive{
		Metadata: &physicaldrive.Metadata{ID: id},
		Vendor:   vendor,
	}
}

// discovererFactory abstracts over NewMegaRAID / NewSmartArray so we can
// run the same table-driven tests against both implementations.
type discovererFactory struct {
	name           string
	controllerType string
	build          func(m *mockRAIDController) service.PhysicalDriveDiscoverer
}

var factories = []discovererFactory{
	{
		name:           "MegaRAID",
		controllerType: megaraidControllerType,
		build:          func(m *mockRAIDController) service.PhysicalDriveDiscoverer { return NewMegaRAID(m) },
	},
	{
		name:           "SmartArray",
		controllerType: smartArrayControllerType,
		build:          func(m *mockRAIDController) service.PhysicalDriveDiscoverer { return NewSmartArray(m) },
	},
}

func TestDiscoverPhysicalDrives_SingleController(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			ctrl0 := newController(0)
			drives := []*physicaldrive.PhysicalDrive{
				newDrive("0:1:0", "Seagate"),
				newDrive("0:1:1", "HGST"),
			}

			rc.On("Controllers").Return([]*raidcontroller.RAIDController{ctrl0}, nil)
			rc.On("PhysicalDrives", ctrl0.Metadata).Return(drives, nil)

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.NoError(t, err)
			require.Len(t, result, 2)
			assert.Equal(t, f.controllerType, result[0].ControllerType)
			assert.Equal(t, 0, result[0].ControllerID)
			assert.Equal(t, "Seagate", result[0].Vendor)
			assert.Equal(t, f.controllerType, result[1].ControllerType)
			assert.Equal(t, "HGST", result[1].Vendor)
			rc.AssertExpectations(t)
		})
	}
}

func TestDiscoverPhysicalDrives_MultipleControllers(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			ctrl0 := newController(0)
			ctrl1 := newController(1)

			rc.On("Controllers").Return([]*raidcontroller.RAIDController{ctrl0, ctrl1}, nil)
			rc.On("PhysicalDrives", ctrl0.Metadata).Return([]*physicaldrive.PhysicalDrive{newDrive("0:1:0", "Seagate")}, nil)
			rc.On("PhysicalDrives", ctrl1.Metadata).Return([]*physicaldrive.PhysicalDrive{newDrive("1:1:0", "Samsung")}, nil)

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.NoError(t, err)
			require.Len(t, result, 2)
			assert.Equal(t, 0, result[0].ControllerID)
			assert.Equal(t, "Seagate", result[0].Vendor)
			assert.Equal(t, 1, result[1].ControllerID)
			assert.Equal(t, "Samsung", result[1].Vendor)
			rc.AssertExpectations(t)
		})
	}
}

func TestDiscoverPhysicalDrives_NoControllers(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			rc.On("Controllers").Return([]*raidcontroller.RAIDController{}, nil)

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.NoError(t, err)
			assert.Empty(t, result)
			rc.AssertExpectations(t)
		})
	}
}

func TestDiscoverPhysicalDrives_NoDrives(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			ctrl0 := newController(0)
			rc.On("Controllers").Return([]*raidcontroller.RAIDController{ctrl0}, nil)
			rc.On("PhysicalDrives", ctrl0.Metadata).Return([]*physicaldrive.PhysicalDrive{}, nil)

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.NoError(t, err)
			assert.Empty(t, result)
			rc.AssertExpectations(t)
		})
	}
}

func TestDiscoverPhysicalDrives_ControllersError(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			rc.On("Controllers").Return(nil, fmt.Errorf("hardware fault"))

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "hardware fault")
			assert.Contains(t, err.Error(), f.name)
			rc.AssertExpectations(t)
		})
	}
}

func TestDiscoverPhysicalDrives_PhysicalDrivesError(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			rc := new(mockRAIDController)
			ctrl0 := newController(0)
			rc.On("Controllers").Return([]*raidcontroller.RAIDController{ctrl0}, nil)
			rc.On("PhysicalDrives", ctrl0.Metadata).Return(nil, fmt.Errorf("I/O timeout"))

			d := f.build(rc)
			result, err := d.DiscoverPhysicalDrives()

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "I/O timeout")
			assert.Contains(t, err.Error(), "controller 0")
			rc.AssertExpectations(t)
		})
	}
}
