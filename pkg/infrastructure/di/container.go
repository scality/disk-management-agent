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
	"os"

	"github.com/go-logr/logr"
	"github.com/scality/raidmgmt/pkg/implementation/commandrunner"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"disk-management-agent/pkg/infrastructure/discovereddrivecache"
	"disk-management-agent/pkg/infrastructure/discoveredphysicaldiskstore"
	"disk-management-agent/pkg/infrastructure/logicalvolumediscoverer"
	"disk-management-agent/pkg/infrastructure/physicaldrivediscoverer"
	"disk-management-agent/pkg/usecase"
)

type Container struct {
	logger    logr.Logger
	k8sClient client.Client
	nodeName  string

	storcliPath string
	perccliPath string
	ssacliPath  string

	// Availability of each vendor CLI on disk, evaluated once at construction
	// time. When a CLI is missing the corresponding discoverers are skipped
	// instead of crashing the controller manager, so that nodes without a
	// given RAID vendor can still run the agent.
	megaraidPerccliAvailable bool
	megaraidStorcliAvailable bool
	smartArrayAvailable      bool

	megaraidPerccliCommandRunner *megaraid.MegaRAIDRunner
	megaraidStorcliCommandRunner *megaraid.MegaRAIDRunner
	ssacliCommandRunner          *commandrunner.SSACLI
	lsblkCommandRunner           *commandrunner.LSBLK

	megaraidPerccliRAIDController *megaraid.Adapter
	megaraidStorcliRAIDController *megaraid.Adapter
	smartArrayRAIDController      *raidcontroller.SmartArray

	megaraidPerccliDiscoverer *physicaldrivediscoverer.MegaRAID
	megaraidStorcliDiscoverer *physicaldrivediscoverer.MegaRAID
	smartArrayDiscoverer      *physicaldrivediscoverer.SmartArray

	megaraidPerccliLVDiscoverer *logicalvolumediscoverer.MegaRAID
	megaraidStorcliLVDiscoverer *logicalvolumediscoverer.MegaRAID
	smartArrayLVDiscoverer      *logicalvolumediscoverer.SmartArray

	discoveredPhysicalDiskStore *discoveredphysicaldiskstore.Kubernetes
	discoveredDriveCache        *discovereddrivecache.InMemory

	discoverPhysicalDrivesUseCase     *usecase.DiscoverPhysicalDrives
	reconcileDiscoveredPhysicalDiskUC *usecase.ReconcileDiscoveredPhysicalDisk
}

func NewContainer(
	logger logr.Logger,
	k8sClient client.Client,
	nodeName string,
	storcliPath string,
	perccliPath string,
	ssacliPath string,
) *Container {
	c := &Container{
		logger:      logger,
		k8sClient:   k8sClient,
		nodeName:    nodeName,
		storcliPath: storcliPath,
		perccliPath: perccliPath,
		ssacliPath:  ssacliPath,
	}

	c.megaraidPerccliAvailable = c.detectBinary("perccli", "MegaRAID PERC", perccliPath)
	c.megaraidStorcliAvailable = c.detectBinary("storcli", "MegaRAID", storcliPath)
	c.smartArrayAvailable = c.detectBinary("ssacli", "HPE Smart Array", ssacliPath)

	return c
}

// detectBinary reports whether the given vendor CLI is usable on this node.
// It logs an informational message when the binary is missing so operators
// understand why a vendor's discovery is disabled. Errors other than
// "not exist" are also logged but treated as unavailable, since we cannot
// safely invoke the binary in that state.
func (c *Container) detectBinary(name, vendor, path string) bool {
	if path == "" {
		c.logger.Info("vendor CLI path is empty, discovery disabled",
			"binary", name, "vendor", vendor)

		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Info("vendor CLI not found, discovery disabled",
				"binary", name, "vendor", vendor, "path", path)
		} else {
			c.logger.Error(err, "failed to stat vendor CLI, discovery disabled",
				"binary", name, "vendor", vendor, "path", path)
		}

		return false
	}

	if info.IsDir() {
		c.logger.Info("vendor CLI path is a directory, discovery disabled",
			"binary", name, "vendor", vendor, "path", path)

		return false
	}

	c.logger.Info("vendor CLI detected, discovery enabled",
		"binary", name, "vendor", vendor, "path", path)

	return true
}
