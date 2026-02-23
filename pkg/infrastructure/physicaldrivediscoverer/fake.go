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
	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"

	"platform-disk-management-agent/pkg/domain"
	"platform-disk-management-agent/pkg/service"
)

// Fake implements PhysicalDriveDiscoverer with hardcoded drives for
// development and testing on systems without RAID hardware (e.g. minikube).
// It simulates a MegaRAID controller with 4 HDDs and 1 SSD.
type Fake struct{}

var _ service.PhysicalDriveDiscoverer = &Fake{}

func NewFake() *Fake {
	return &Fake{}
}

func (*Fake) DiscoverPhysicalDrives() ([]*domain.DiscoveredPhysicalDrive, error) {
	ctrlMeta := &raidcontroller.Metadata{ID: 0}

	return []*domain.DiscoveredPhysicalDrive{
		{
			ControllerType: "MegaRAID",
			ControllerID:   0,
			PhysicalDrive: &physicaldrive.PhysicalDrive{
				Metadata:      &physicaldrive.Metadata{CtrlMetadata: ctrlMeta, ID: "252:0"},
				Slot:          &physicaldrive.Slot{Enclosure: "252", Bay: "0"},
				Vendor:        "Seagate",
				Model:         "ST4000NM0033",
				Serial:        "FAKE0001",
				WWN:           "5000C50000000001",
				Size:          4000787030016,
				Type:          physicaldrive.DiskTypeHDD,
				Status:        physicaldrive.PDStatusUsed,
				DevicePath:    "/dev/sda",
				PermanentPath: "/dev/disk/by-id/wwn-0x5000c50000000001",
			},
		},
		{
			ControllerType: "MegaRAID",
			ControllerID:   0,
			PhysicalDrive: &physicaldrive.PhysicalDrive{
				Metadata:      &physicaldrive.Metadata{CtrlMetadata: ctrlMeta, ID: "252:1"},
				Slot:          &physicaldrive.Slot{Enclosure: "252", Bay: "1"},
				Vendor:        "Seagate",
				Model:         "ST4000NM0033",
				Serial:        "FAKE0002",
				WWN:           "5000C50000000002",
				Size:          4000787030016,
				Type:          physicaldrive.DiskTypeHDD,
				Status:        physicaldrive.PDStatusUnassignedGood,
				DevicePath:    "/dev/sdb",
				PermanentPath: "/dev/disk/by-id/wwn-0x5000c50000000002",
			},
		},
		{
			ControllerType: "MegaRAID",
			ControllerID:   0,
			PhysicalDrive: &physicaldrive.PhysicalDrive{
				Metadata:      &physicaldrive.Metadata{CtrlMetadata: ctrlMeta, ID: "252:2"},
				Slot:          &physicaldrive.Slot{Enclosure: "252", Bay: "2"},
				Vendor:        "HGST",
				Model:         "HUS726040ALA610",
				Serial:        "FAKE0003",
				WWN:           "5000C50000000003",
				Size:          4000787030016,
				Type:          physicaldrive.DiskTypeHDD,
				Status:        physicaldrive.PDStatusUsed,
				DevicePath:    "/dev/sdc",
				PermanentPath: "/dev/disk/by-id/wwn-0x5000c50000000003",
			},
		},
		{
			ControllerType: "MegaRAID",
			ControllerID:   0,
			PhysicalDrive: &physicaldrive.PhysicalDrive{
				Metadata: &physicaldrive.Metadata{CtrlMetadata: ctrlMeta, ID: "252:3"},
				Slot:     &physicaldrive.Slot{Enclosure: "252", Bay: "3"},
				Vendor:   "Seagate",
				Model:    "ST2000NM0008",
				Serial:   "FAKE0004",
				WWN:      "5000C50000000004",
				Size:     2000398934016,
				Type:     physicaldrive.DiskTypeHDD,
				Status:   physicaldrive.PDStatusFailed,
				Reason:   "Predictive failure",
			},
		},
		{
			ControllerType: "MegaRAID",
			ControllerID:   0,
			PhysicalDrive: &physicaldrive.PhysicalDrive{
				Metadata:      &physicaldrive.Metadata{CtrlMetadata: ctrlMeta, ID: "252:4"},
				Slot:          &physicaldrive.Slot{Enclosure: "252", Bay: "4"},
				Vendor:        "Samsung",
				Model:         "MZ7LH480HAHQ",
				Serial:        "FAKE0005",
				WWN:           "5000C50000000005",
				Size:          480103981056,
				Type:          physicaldrive.DiskTypeSSD,
				Status:        physicaldrive.PDStatusUsed,
				JBOD:          true,
				DevicePath:    "/dev/sde",
				PermanentPath: "/dev/disk/by-id/wwn-0x5000c50000000005",
			},
		},
	}, nil
}
