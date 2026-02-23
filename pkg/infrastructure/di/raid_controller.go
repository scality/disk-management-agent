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

package di

import (
	"github.com/scality/raidmgmt/pkg/implementation/blinker"
	"github.com/scality/raidmgmt/pkg/implementation/controllergetter"
	"github.com/scality/raidmgmt/pkg/implementation/logicalvolumegetter"
	"github.com/scality/raidmgmt/pkg/implementation/logicalvolumemanager"
	"github.com/scality/raidmgmt/pkg/implementation/physicaldrivegetter"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
)

func (c *Container) getMegaRAIDPerccliRAIDController() *megaraid.Adapter {
	if c.megaraidPerccliRAIDController == nil {
		c.megaraidPerccliRAIDController = megaraid.New(c.getMegaRAIDPerccliCommandRunner())
	}

	return c.megaraidPerccliRAIDController
}

func (c *Container) getMegaRAIDStorcliRAIDController() *megaraid.Adapter {
	if c.megaraidStorcliRAIDController == nil {
		c.megaraidStorcliRAIDController = megaraid.New(c.getMegaRAIDStorcliCommandRunner())
	}

	return c.megaraidStorcliRAIDController
}

func (c *Container) getSmartArrayRAIDController() *raidcontroller.SmartArray {
	if c.smartArrayRAIDController == nil {
		ssacliRunner := c.getSSACLICommandRunner()
		lsblkRunner := c.getLSBLKCommandRunner()

		ctrlGetter := controllergetter.NewSSACLI(ssacliRunner)
		pdGetter := physicaldrivegetter.NewSSACLI(ssacliRunner, lsblkRunner)
		lvGetter := logicalvolumegetter.NewSSACLI(ssacliRunner, lsblkRunner)
		lvManager := logicalvolumemanager.NewSSACLI(ssacliRunner, pdGetter, lvGetter)
		blk := blinker.NewSSACLI(ssacliRunner)

		c.smartArrayRAIDController = raidcontroller.NewSmartArray(
			ctrlGetter,
			pdGetter,
			lvGetter,
			lvManager,
			blk,
		)
	}

	return c.smartArrayRAIDController
}
