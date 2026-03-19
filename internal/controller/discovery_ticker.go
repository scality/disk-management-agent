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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"disk-management-agent/pkg/usecase"
)

// DefaultDiscoveryInterval is the default interval at which the discovery
// ticker fires.
const DefaultDiscoveryInterval = 5 * time.Minute

// DiscoveryTicker periodically triggers disk discovery via the
// DiscoverPhysicalDrives use case and sends GenericEvents for existing
// CRs so the reconciler can update them.
// It implements manager.Runnable (Start(context.Context) error) so it can
// be registered with the controller-runtime manager via mgr.Add().
type DiscoveryTicker struct {
	// NodeName is the name of the Kubernetes node this agent is running on.
	NodeName string
	// Interval is the duration between consecutive discovery ticks.
	Interval time.Duration
	// EventChan is a send-only channel used to push GenericEvents that
	// trigger reconciliation of the DiscoveredPhysicalDisk controller.
	EventChan chan<- event.GenericEvent
	// UseCase performs the actual drive discovery and CR creation.
	UseCase *usecase.DiscoverPhysicalDrives
}

// Start runs the ticker loop until the context is cancelled.
// It satisfies the manager.Runnable interface.
func (t *DiscoveryTicker) Start(ctx context.Context) error {
	logger := logf.FromContext(ctx).WithName("discovery-ticker")
	logger.Info("Starting discovery ticker", "interval", t.Interval, "nodeName", t.NodeName)

	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping discovery ticker")
			return nil
		case <-ticker.C:
			logger.Info("Discovery tick")
			t.discover(ctx)
		}
	}
}

func (t *DiscoveryTicker) discover(ctx context.Context) {
	logger := logf.FromContext(ctx).WithName("discovery-ticker")

	existingNames, err := t.UseCase.Execute(ctx)
	if err != nil {
		logger.Error(err, "Discovery use case failed")
		return
	}

	for _, name := range existingNames {
		t.EventChan <- event.GenericEvent{
			Object: &metav1.PartialObjectMetadata{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			},
		}
	}

	logger.Info("Discovery tick completed",
		"existingCRs", len(existingNames),
	)
}
