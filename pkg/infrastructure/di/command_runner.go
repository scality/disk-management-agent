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
	"github.com/scality/raidmgmt/pkg/implementation/commandrunner"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
)

func (c *Container) getMegaRAIDPerccliCommandRunner() *megaraid.MegaRAIDRunner {
	if !c.megaraidPerccliAvailable {
		return nil
	}

	if c.megaraidPerccliCommandRunner == nil {
		runner, err := megaraid.NewMegaRAIDRunner(c.perccliPath)
		if err != nil {
			c.logger.Error(err, "Failed to create MegaRAID perccli runner, discovery disabled",
				"path", c.perccliPath)
			c.megaraidPerccliAvailable = false

			return nil
		}

		c.megaraidPerccliCommandRunner = runner
	}

	return c.megaraidPerccliCommandRunner
}

func (c *Container) getMegaRAIDStorcliCommandRunner() *megaraid.MegaRAIDRunner {
	if !c.megaraidStorcliAvailable {
		return nil
	}

	if c.megaraidStorcliCommandRunner == nil {
		runner, err := megaraid.NewMegaRAIDRunner(c.storcliPath)
		if err != nil {
			c.logger.Error(err, "Failed to create MegaRAID storcli runner, discovery disabled",
				"path", c.storcliPath)
			c.megaraidStorcliAvailable = false

			return nil
		}

		c.megaraidStorcliCommandRunner = runner
	}

	return c.megaraidStorcliCommandRunner
}

func (c *Container) getSSACLICommandRunner() *commandrunner.SSACLI {
	if !c.smartArrayAvailable {
		return nil
	}

	if c.ssacliCommandRunner == nil {
		c.ssacliCommandRunner = commandrunner.NewSSACLI(&c.ssacliPath)
	}

	return c.ssacliCommandRunner
}

func (c *Container) getLSBLKCommandRunner() *commandrunner.LSBLK {
	if c.lsblkCommandRunner == nil {
		c.lsblkCommandRunner = commandrunner.NewLSBLK(nil)
	}

	return c.lsblkCommandRunner
}
