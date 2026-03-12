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
	"github.com/pkg/errors"
	"github.com/scality/raidmgmt/pkg/domain/ports"

	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

const smartArrayControllerType = "SmartArray"

// SmartArray discovers physical drives behind HPE Smart Array controllers
// using ssacli.
type SmartArray struct {
	rc ports.RAIDController
}

var _ service.PhysicalDriveDiscoverer = &SmartArray{}

func NewSmartArray(rc ports.RAIDController) *SmartArray {
	return &SmartArray{rc: rc}
}

func (d *SmartArray) DiscoverPhysicalDrives() ([]*domain.DiscoveredPhysicalDrive, error) {
	controllers, err := d.rc.Controllers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list SmartArray controllers")
	}

	var drives []*domain.DiscoveredPhysicalDrive

	for _, ctrl := range controllers {
		pds, err := d.rc.PhysicalDrives(ctrl.Metadata)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list physical drives for SmartArray controller %d", ctrl.Metadata.ID)
		}

		for _, pd := range pds {
			drives = append(drives, &domain.DiscoveredPhysicalDrive{
				ControllerType: smartArrayControllerType,
				ControllerID:   ctrl.Metadata.ID,
				PhysicalDrive:  pd,
			})
		}
	}

	return drives, nil
}
