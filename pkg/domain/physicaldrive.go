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

package domain

import (
	"fmt"
	"strings"

	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
)

// DiscoveredPhysicalDrive extends the raidmgmt PhysicalDrive with RAID
// controller type context needed for CR identification.
type DiscoveredPhysicalDrive struct {
	ControllerType string
	ControllerID   int
	*physicaldrive.PhysicalDrive
}

// ComputeCRName builds the Kubernetes resource name for a discovered
// physical drive following the convention:
//
//	<node-name>-<controller-type>-<controller-id>-<slot-str>
//
// where slot-str is the drive's slot identifier with colons removed and
// lowercased to produce a valid DNS subdomain label.
func ComputeCRName(nodeName, controllerType string, controllerID int, slotID string) string {
	sanitized := strings.ToLower(strings.ReplaceAll(slotID, ":", ""))
	return fmt.Sprintf("%s-%s-%d-%s", nodeName, strings.ToLower(controllerType), controllerID, sanitized)
}
