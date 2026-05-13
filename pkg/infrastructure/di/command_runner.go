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
	"os/exec"

	"github.com/scality/raidmgmt/pkg/implementation/commandrunner"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
)

func (c *Container) getMegaRAIDPerccliCommandRunner() *megaraid.MegaRAIDRunner {
	if c.megaraidPerccliCommandRunner != nil {
		return c.megaraidPerccliCommandRunner
	}

	if c.megaraidPerccliCommandRunnerTried {
		return nil
	}

	c.megaraidPerccliCommandRunnerTried = true

	runner, err := megaraid.NewMegaRAIDRunner(c.perccliPath)
	if err != nil {
		c.logger.Info(
			"MegaRAID perccli runner unavailable, related features will be disabled",
			"path", c.perccliPath,
			"error", err.Error(),
		)

		return nil
	}

	c.megaraidPerccliCommandRunner = runner

	return c.megaraidPerccliCommandRunner
}

func (c *Container) getMegaRAIDStorcliCommandRunner() *megaraid.MegaRAIDRunner {
	if c.megaraidStorcliCommandRunner != nil {
		return c.megaraidStorcliCommandRunner
	}

	if c.megaraidStorcliCommandRunnerTried {
		return nil
	}

	c.megaraidStorcliCommandRunnerTried = true

	runner, err := megaraid.NewMegaRAIDRunner(c.storcliPath)
	if err != nil {
		c.logger.Info(
			"MegaRAID storcli runner unavailable, related features will be disabled",
			"path", c.storcliPath,
			"error", err.Error(),
		)

		return nil
	}

	c.megaraidStorcliCommandRunner = runner

	return c.megaraidStorcliCommandRunner
}

func (c *Container) getSSACLICommandRunner() *commandrunner.SSACLI {
	if c.ssacliCommandRunner != nil {
		return c.ssacliCommandRunner
	}

	if c.ssacliCommandRunnerTried {
		return nil
	}

	c.ssacliCommandRunnerTried = true

	if _, err := exec.LookPath(c.ssacliPath); err != nil {
		c.logger.Info(
			"ssacli runner unavailable, related features will be disabled",
			"path", c.ssacliPath,
			"error", err.Error(),
		)

		return nil
	}

	c.ssacliCommandRunner = commandrunner.NewSSACLI(&c.ssacliPath)

	return c.ssacliCommandRunner
}

func (c *Container) getLSBLKCommandRunner() *commandrunner.LSBLK {
	if c.lsblkCommandRunner == nil {
		c.lsblkCommandRunner = commandrunner.NewLSBLK(nil)
	}

	return c.lsblkCommandRunner
}
