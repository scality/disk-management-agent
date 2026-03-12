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
	"github.com/go-logr/logr"
	"github.com/scality/raidmgmt/pkg/implementation/commandrunner"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"disk-management-agent/pkg/infrastructure/discovereddrivecache"
	"disk-management-agent/pkg/infrastructure/discoveredphysicaldiskstore"
	"disk-management-agent/pkg/infrastructure/physicaldrivediscoverer"
	"disk-management-agent/pkg/usecase"
)

type Container struct {
	logger    logr.Logger
	k8sClient client.Client
	nodeName  string
	namespace string

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

	discoveredPhysicalDiskStore *discoveredphysicaldiskstore.Kubernetes
	discoveredDriveCache        *discovereddrivecache.InMemory

	discoverPhysicalDrivesUseCase     *usecase.DiscoverPhysicalDrives
	reconcileDiscoveredPhysicalDiskUC *usecase.ReconcileDiscoveredPhysicalDisk
}

func NewContainer(
	logger logr.Logger,
	k8sClient client.Client,
	nodeName string,
	namespace string,
) *Container {
	return &Container{
		logger:    logger,
		k8sClient: k8sClient,
		nodeName:  nodeName,
		namespace: namespace,
	}
}
