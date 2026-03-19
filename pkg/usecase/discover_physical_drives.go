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

package usecase

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/service"
)

// DiscoverPhysicalDrives orchestrates physical drive discovery across all
// RAID controllers, filters out non-HDD drives, and ensures each drive
// has a corresponding DiscoveredPhysicalDisk CR in Kubernetes.
type DiscoverPhysicalDrives struct {
	logger      logr.Logger
	discoverers []service.PhysicalDriveDiscoverer
	store       service.DiscoveredPhysicalDiskStore
	nodeName    string
}

// NewDiscoverPhysicalDrives creates a new DiscoverPhysicalDrives use case.
func NewDiscoverPhysicalDrives(
	logger logr.Logger,
	discoverers []service.PhysicalDriveDiscoverer,
	store service.DiscoveredPhysicalDiskStore,
	nodeName string,
) *DiscoverPhysicalDrives {
	return &DiscoverPhysicalDrives{
		logger:      logger.WithName("discover-physical-drives"),
		discoverers: discoverers,
		store:       store,
		nodeName:    nodeName,
	}
}

// Execute runs discovery across all registered RAID controller adapters and
// synchronizes the results with Kubernetes.
//
// It returns the names of CRs that already existed (and need a reconcile
// trigger). Newly created CRs are handled automatically by the controller's
// watch.
func (u *DiscoverPhysicalDrives) Execute(ctx context.Context) ([]string, error) {
	allDrives := u.gatherDrives()

	var existingCRNames []string

	for _, drive := range allDrives {
		if drive.Type != physicaldrive.DiskTypeHDD {
			u.logger.V(1).Info(
				"Skipping non-HDD drive",
				"type", drive.Type.String(),
				"slot", drive.Slot.Format(),
				"controllerType", drive.ControllerType,
				"controllerID", drive.ControllerID,
			)

			continue
		}

		crName := domain.ComputeCRName(
			u.nodeName,
			drive.ControllerType,
			drive.ControllerID,
			drive.ID,
		)

		existing, err := u.store.Get(ctx, crName)
		if err != nil {
			u.logger.Error(err, "Failed to check DiscoveredPhysicalDisk existence", "name", crName)

			continue
		}

		if existing != nil {
			u.logger.V(1).Info("DiscoveredPhysicalDisk already exists, queuing for reconcile", "name", crName)
			existingCRNames = append(existingCRNames, crName)

			continue
		}

		cr := buildCR(crName, u.nodeName, drive)

		if err := u.store.Create(ctx, cr); err != nil {
			u.logger.Error(err, "Failed to create DiscoveredPhysicalDisk", "name", crName)

			continue
		}

		u.logger.Info("Created DiscoveredPhysicalDisk", "name", crName)
	}

	return existingCRNames, nil
}

// gatherDrives collects physical drives from all registered discoverers.
// Errors from individual discoverers are logged and skipped so that a
// missing CLI tool (e.g. storcli on a SmartArray-only host) does not
// block discovery for other controller types.
func (u *DiscoverPhysicalDrives) gatherDrives() []*domain.DiscoveredPhysicalDrive {
	var allDrives []*domain.DiscoveredPhysicalDrive

	for _, discoverer := range u.discoverers {
		drives, err := discoverer.DiscoverPhysicalDrives()
		if err != nil {
			u.logger.V(1).Info("Discoverer returned an error, skipping", "error", err)

			continue
		}

		allDrives = append(allDrives, drives...)
	}

	return allDrives
}

func buildCR(
	name, nodeName string,
	drive *domain.DiscoveredPhysicalDrive,
) *metalk8sv1alpha1.DiscoveredPhysicalDisk {
	slot := metalk8sv1alpha1.SlotLocation{}
	if drive.Slot != nil {
		slot.Port = drive.Slot.Port
		slot.Enclosure = drive.Slot.Enclosure
		slot.Bay = drive.Slot.Bay
	}

	return &metalk8sv1alpha1.DiscoveredPhysicalDisk{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: metalk8sv1alpha1.DiscoveredPhysicalDiskSpec{
			NodeName: nodeName,
			Controller: metalk8sv1alpha1.ControllerRef{
				Type: drive.ControllerType,
				ID:   drive.ControllerID,
			},
			ID:     drive.ID,
			Slot:   slot,
			Vendor: &drive.Vendor,
			Model:  &drive.Model,
			Serial: &drive.Serial,
			WWN:    &drive.WWN,
			Size:   int64(drive.Size), //nolint:gosec // raidmgmt uses uint64 for size; overflow is not a concern for disk sizes
			Type:   drive.Type.String(),
		},
	}
}
