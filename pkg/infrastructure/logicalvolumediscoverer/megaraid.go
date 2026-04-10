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

//nolint:dupl // This function is similar to the one in the SmartArray implementation.
package logicalvolumediscoverer

import (
	"github.com/scality/go-errors"
	"github.com/scality/raidmgmt/pkg/domain/ports"

	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

const megaraidControllerType = "MegaRAID"

// MegaRAID discovers logical volumes behind MegaRAID/PERC controllers
// using storcli or perccli.
type MegaRAID struct {
	rc ports.RAIDController
}

var _ service.LogicalVolumeDiscoverer = &MegaRAID{}

func NewMegaRAID(rc ports.RAIDController) *MegaRAID {
	return &MegaRAID{rc: rc}
}

func (d *MegaRAID) DiscoverLogicalVolumes() ([]*domain.DiscoveredLogicalVolume, error) {
	controllers, err := d.rc.Controllers()
	if err != nil {
		return nil, errors.Wrap(domain.ErrControllerDiscovery,
			errors.WithDetail("failed to list controllers"),
			errors.WithProperty("controller_type", megaraidControllerType),
			errors.CausedBy(err),
		)
	}

	var volumes []*domain.DiscoveredLogicalVolume

	for _, ctrl := range controllers {
		lvs, err := d.rc.LogicalVolumes(ctrl.Metadata)
		if err != nil {
			return nil, errors.Wrap(domain.ErrVolumeDiscovery,
				errors.WithDetail("failed to list logical volumes from controller"),
				errors.WithProperty("controller_id", ctrl.ID),
				errors.WithProperty("controller_type", megaraidControllerType),
				errors.CausedBy(err),
			)
		}

		for _, lv := range lvs {
			volumes = append(volumes, &domain.DiscoveredLogicalVolume{
				ControllerType: megaraidControllerType,
				ControllerID:   ctrl.ID,
				LogicalVolume:  lv,
			})
		}
	}

	return volumes, nil
}
