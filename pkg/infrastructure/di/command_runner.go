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

// getStorcli2CommandRunner returns the storcli2 command runner, or nil if the
// binary is absent. Unlike NewMegaRAIDRunner, commandrunner.NewStorCLI2 does
// not validate the binary path at construction, so availability is probed with
// exec.LookPath (as for ssacli) to keep the warn-and-skip behaviour.
func (c *Container) getStorcli2CommandRunner() *commandrunner.StorCLI2 {
	if c.storcli2CommandRunner != nil {
		return c.storcli2CommandRunner
	}

	if c.storcli2CommandRunnerTried {
		return nil
	}

	c.storcli2CommandRunnerTried = true

	if _, err := exec.LookPath(c.storcli2Path); err != nil {
		c.logger.Info(
			"storcli2 runner unavailable, related features will be disabled",
			"path", c.storcli2Path,
			"error", err.Error(),
		)

		return nil
	}

	c.storcli2CommandRunner = commandrunner.NewStorCLI2(&c.storcli2Path)

	return c.storcli2CommandRunner
}

// getPerccli2CommandRunner mirrors getStorcli2CommandRunner for perccli2.
func (c *Container) getPerccli2CommandRunner() *commandrunner.PercCLI2 {
	if c.perccli2CommandRunner != nil {
		return c.perccli2CommandRunner
	}

	if c.perccli2CommandRunnerTried {
		return nil
	}

	c.perccli2CommandRunnerTried = true

	if _, err := exec.LookPath(c.perccli2Path); err != nil {
		c.logger.Info(
			"perccli2 runner unavailable, related features will be disabled",
			"path", c.perccli2Path,
			"error", err.Error(),
		)

		return nil
	}

	c.perccli2CommandRunner = commandrunner.NewPercCLI2(&c.perccli2Path)

	return c.perccli2CommandRunner
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
