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

//nolint:dupl // This function is similar to the one in the MegaRAID implementation.
package logicalvolumediscoverer

import (
	"github.com/pkg/errors"
	"github.com/scality/raidmgmt/pkg/domain/ports"

	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

const smartArrayControllerType = "SmartArray"

// SmartArray discovers logical volumes behind HPE Smart Array controllers
// using ssacli.
type SmartArray struct {
	rc ports.RAIDController
}

var _ service.LogicalVolumeDiscoverer = &SmartArray{}

func NewSmartArray(rc ports.RAIDController) *SmartArray {
	return &SmartArray{rc: rc}
}

func (d *SmartArray) DiscoverLogicalVolumes() ([]*domain.DiscoveredLogicalVolume, error) {
	controllers, err := d.rc.Controllers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list SmartArray controllers")
	}

	var volumes []*domain.DiscoveredLogicalVolume

	for _, ctrl := range controllers {
		lvs, err := d.rc.LogicalVolumes(ctrl.Metadata)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list logical volumes for SmartArray controller %d", ctrl.ID)
		}

		for _, lv := range lvs {
			volumes = append(volumes, &domain.DiscoveredLogicalVolume{
				ControllerType: smartArrayControllerType,
				ControllerID:   ctrl.ID,
				LogicalVolume:  lv,
			})
		}
	}

	return volumes, nil
}
