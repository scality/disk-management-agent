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

//nolint:dupl // Mirrors physical_drive_discoverer.go (same vendor dispatch, different discoverer type).
package di

import (
	"github.com/scality/raidmgmt/pkg/core"

	"disk-management-agent/pkg/infrastructure/logicalvolumediscoverer"
)

func (c *Container) getMegaRAIDPerccliLVDiscoverer() *logicalvolumediscoverer.MegaRAID {
	if c.megaraidPerccliLVDiscoverer == nil {
		ctrl := c.getMegaRAIDPerccliRAIDController()
		if ctrl == nil {
			return nil
		}

		c.megaraidPerccliLVDiscoverer = logicalvolumediscoverer.NewMegaRAID(
			core.NewRAIDController(ctrl),
		)
	}

	return c.megaraidPerccliLVDiscoverer
}

func (c *Container) getMegaRAIDStorcliLVDiscoverer() *logicalvolumediscoverer.MegaRAID {
	if c.megaraidStorcliLVDiscoverer == nil {
		ctrl := c.getMegaRAIDStorcliRAIDController()
		if ctrl == nil {
			return nil
		}

		c.megaraidStorcliLVDiscoverer = logicalvolumediscoverer.NewMegaRAID(
			core.NewRAIDController(ctrl),
		)
	}

	return c.megaraidStorcliLVDiscoverer
}

func (c *Container) getSmartArrayLVDiscoverer() *logicalvolumediscoverer.SmartArray {
	if c.smartArrayLVDiscoverer == nil {
		ctrl := c.getSmartArrayRAIDController()
		if ctrl == nil {
			return nil
		}

		c.smartArrayLVDiscoverer = logicalvolumediscoverer.NewSmartArray(
			core.NewRAIDController(ctrl),
		)
	}

	return c.smartArrayLVDiscoverer
}
