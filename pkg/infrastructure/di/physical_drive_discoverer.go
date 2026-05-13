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

//nolint:dupl // The logical-volume discoverer wiring mirrors this file by design.
package di

import (
	"github.com/scality/raidmgmt/pkg/core"

	"disk-management-agent/pkg/infrastructure/physicaldrivediscoverer"
)

func (c *Container) getMegaRAIDPerccliDiscoverer() *physicaldrivediscoverer.MegaRAID {
	if c.megaraidPerccliDiscoverer == nil {
		rc := c.getMegaRAIDPerccliRAIDController()
		if rc == nil {
			return nil
		}

		c.megaraidPerccliDiscoverer = physicaldrivediscoverer.NewMegaRAID(
			core.NewRAIDController(rc),
		)
	}

	return c.megaraidPerccliDiscoverer
}

func (c *Container) getMegaRAIDStorcliDiscoverer() *physicaldrivediscoverer.MegaRAID {
	if c.megaraidStorcliDiscoverer == nil {
		rc := c.getMegaRAIDStorcliRAIDController()
		if rc == nil {
			return nil
		}

		c.megaraidStorcliDiscoverer = physicaldrivediscoverer.NewMegaRAID(
			core.NewRAIDController(rc),
		)
	}

	return c.megaraidStorcliDiscoverer
}

func (c *Container) getSmartArrayDiscoverer() *physicaldrivediscoverer.SmartArray {
	if c.smartArrayDiscoverer == nil {
		rc := c.getSmartArrayRAIDController()
		if rc == nil {
			return nil
		}

		c.smartArrayDiscoverer = physicaldrivediscoverer.NewSmartArray(
			core.NewRAIDController(rc),
		)
	}

	return c.smartArrayDiscoverer
}
