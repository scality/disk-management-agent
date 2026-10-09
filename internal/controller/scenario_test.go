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

package controller

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/scality/raidmgmt/pkg/core"
	"github.com/scality/raidmgmt/pkg/domain/ports"
	"github.com/scality/raidmgmt/pkg/implementation/controllergetter"
	"github.com/scality/raidmgmt/pkg/implementation/logicalvolumegetter"
	"github.com/scality/raidmgmt/pkg/implementation/physicaldrivegetter"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller"
	"github.com/scality/raidmgmt/pkg/implementation/raidcontroller/megaraid"
	"github.com/scality/raidmgmt/scenario"
	"github.com/scality/raidmgmt/scenario/megaraidsim"
	"github.com/scality/raidmgmt/scenario/ssaclisim"
	"github.com/scality/raidmgmt/scenario/storcli2sim"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
	"disk-management-agent/pkg/infrastructure/discovereddrivecache"
	"disk-management-agent/pkg/infrastructure/discoveredphysicaldiskstore"
	"disk-management-agent/pkg/infrastructure/logicalvolumediscoverer"
	"disk-management-agent/pkg/infrastructure/physicaldrivediscoverer"
	"disk-management-agent/pkg/service"
	"disk-management-agent/pkg/usecase"
)

// Hardware scenarios from github.com/scality/raidmgmt/scenario: each phase
// replays real controller output with some drives and volumes changed, runs
// discovery and reconcile as the agent does, and checks the
// DiscoveredPhysicalDisk status read by the Storage Service UI against the
// table of the phase. The scenarios are written in YAML in that module.
// scenarioController builds, for a phase of a scenario, the raidmgmt adapter
// the agent uses, replaying the controller output, and a function restoring
// the host.
type scenarioController func(s *scenario.Scenario, phase int) (ports.RAIDController, func(), error)

// scenarioDiscoverers returns the discoverers the agent uses for an adapter.
type scenarioDiscoverers func(rc *core.RAIDController) (
	service.PhysicalDriveDiscoverer, service.LogicalVolumeDiscoverer,
)

// megaraidDiscoverers are used for every MegaRAID adapter (storcli64,
// perccli64, storcli2, perccli2).
func megaraidDiscoverers(rc *core.RAIDController) (service.PhysicalDriveDiscoverer, service.LogicalVolumeDiscoverer) {
	return physicaldrivediscoverer.NewMegaRAID(rc), logicalvolumediscoverer.NewMegaRAID(rc)
}

// scenarioControllers are the controller families the scenarios cover.
var scenarioControllers = []struct {
	controller  string
	build       scenarioController
	discoverers scenarioDiscoverers
}{
	// storcli64/perccli64: the host is read for the by-id links of volumes.
	{"megaraid", func(s *scenario.Scenario, phase int) (ports.RAIDController, func(), error) {
		ctrl, err := megaraidsim.New(s, phase)
		if err != nil {
			return nil, nil, err
		}

		return megaraid.New(ctrl), ctrl.UseHost(), nil
	}, megaraidDiscoverers},
	// storcli2/perccli2: volume paths come from the controller output only.
	{"storcli2", func(s *scenario.Scenario, phase int) (ports.RAIDController, func(), error) {
		ctrl, err := storcli2sim.New(s, phase)
		if err != nil {
			return nil, nil, err
		}

		return raidcontroller.NewStorCLI2(ctrl), func() {}, nil
	}, megaraidDiscoverers},
	// ssacli (HPE Smart Array): the getters are composed as in the agent's
	// container; only discovery is played, so no manager nor blinker.
	{"ssacli", func(s *scenario.Scenario, phase int) (ports.RAIDController, func(), error) {
		ctrl, err := ssaclisim.New(s, phase)
		if err != nil {
			return nil, nil, err
		}

		return raidcontroller.NewSmartArray(
			&controllergetter.SSACLI{SSACLI: ctrl.SSACLI()},
			&physicaldrivegetter.SSACLI{SSACLI: ctrl.SSACLI(), LSBLK: ctrl.LSBLK()},
			&logicalvolumegetter.SSACLI{SSACLI: ctrl.SSACLI(), LSBLK: ctrl.LSBLK()},
			nil, nil,
		), func() {}, nil
	}, func(rc *core.RAIDController) (service.PhysicalDriveDiscoverer, service.LogicalVolumeDiscoverer) {
		return physicaldrivediscoverer.NewSmartArray(rc), logicalvolumediscoverer.NewSmartArray(rc)
	}},
}

var _ = Describe("Hardware scenarios", Label("scenario"), func() {
	for _, c := range scenarioControllers {
		names, err := scenario.List(c.controller)
		if err != nil {
			panic(err)
		}

		for _, name := range names {
			describeScenario(name, c.build, c.discoverers)
		}
	}
})

// describeScenario plays the phases of a scenario, in order.
func describeScenario(name string, build scenarioController, discoverers scenarioDiscoverers) {
	s, err := scenario.Load(name)
	if err != nil {
		panic(err)
	}

	Describe(s.Name, Ordered, ContinueOnFailure, func() {
		// The cache outlives the phases, as in the agent.
		cache := discovereddrivecache.NewInMemory()
		nodeName := "scenario-" + name

		AfterAll(func(ctx SpecContext) {
			list := &metalk8sv1alpha1.DiscoveredPhysicalDiskList{}
			Expect(k8sClient.List(ctx, list)).To(Succeed())

			for i := range list.Items {
				if list.Items[i].Spec.NodeName == nodeName {
					Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &list.Items[i]))).To(Succeed())
				}
			}
		})

		for i, phase := range s.Phases {
			It(phase.Name, func(ctx SpecContext) {
				rc, restoreHost, err := build(s, i)
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(restoreHost)

				disks := discoverAndReconcile(ctx, rc, discoverers, cache, nodeName)

				notes, err := phase.Check(disks)
				Expect(err).NotTo(HaveOccurred())

				if len(notes) > 0 {
					AddReportEntry("known bug "+phase.KnownBug.Ticket+" reproduced, differences with the expected table",
						strings.Join(notes, "\n"))
				}
			})
		}
	})
}

// discoverAndReconcile runs one discovery tick on a replayed RAID controller,
// reconciles the DiscoveredPhysicalDisks of the node, and returns their status
// as scenario disks.
func discoverAndReconcile(
	ctx context.Context, adapter ports.RAIDController, discoverers scenarioDiscoverers,
	cache *discovereddrivecache.InMemory, nodeName string,
) []scenario.Disk {
	pdDiscoverer, lvDiscoverer := discoverers(core.NewRAIDController(adapter))

	discovery := usecase.NewDiscoverPhysicalDrives(
		logr.Discard(),
		[]service.PhysicalDriveDiscoverer{pdDiscoverer},
		[]service.LogicalVolumeDiscoverer{lvDiscoverer},
		discoveredphysicaldiskstore.NewKubernetes(k8sClient),
		cache,
		nodeName,
	)

	_, err := discovery.Execute(ctx)
	Expect(err).NotTo(HaveOccurred())

	reconciler := &DiscoveredPhysicalDiskReconciler{
		Client:           k8sClient,
		Scheme:           k8sClient.Scheme(),
		NodeName:         nodeName,
		ReconcileUseCase: usecase.NewReconcileDiscoveredPhysicalDisk(cache),
	}

	list := &metalk8sv1alpha1.DiscoveredPhysicalDiskList{}
	Expect(k8sClient.List(ctx, list)).To(Succeed())

	disks := make([]scenario.Disk, 0, len(list.Items))

	for _, item := range list.Items {
		if item.Spec.NodeName != nodeName {
			continue
		}

		key := types.NamespacedName{Name: item.Name}

		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: key})
		Expect(err).NotTo(HaveOccurred())

		disk := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
		Expect(k8sClient.Get(ctx, key, disk)).To(Succeed())

		if disk.Status.Available == nil || !*disk.Status.Available {
			continue
		}

		disks = append(disks, scenario.Disk{
			Slot:          disk.Spec.ID,
			Status:        deref(disk.Status.Status),
			Serial:        deref(disk.Status.Serial),
			DevicePath:    deref(disk.Status.DevicePath),
			PermanentPath: deref(disk.Status.PermanentPath),
		})
	}

	return disks
}

func deref(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
