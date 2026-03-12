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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControllerRef identifies the RAID controller managing this disk.
type ControllerRef struct {
	// Type is the controller type (e.g. "MegaRAID").
	Type string `json:"type"`
	// ID is the controller index.
	ID int `json:"id"`
}

// SlotLocation identifies the physical slot of the disk.
type SlotLocation struct {
	// Port is the port number.
	Port string `json:"port"`
	// Enclosure is the enclosure number.
	Enclosure string `json:"enclosure"`
	// Bay is the bay number.
	Bay string `json:"bay"`
}

// DiscoveredPhysicalDiskSpec defines the desired state of DiscoveredPhysicalDisk.
// It contains only immutable slot identifiers set at creation time.
type DiscoveredPhysicalDiskSpec struct {
	// NodeName is the name of the node where this disk was discovered.
	NodeName string `json:"nodeName"`
	// Controller identifies the RAID controller managing this disk.
	Controller ControllerRef `json:"controller"`
	// ID is the disk identifier as reported by the controller (e.g. "0:1:2").
	ID string `json:"id"`
	// Slot describes the physical slot location of the disk.
	Slot SlotLocation `json:"slot"`
}

// DiscoveredPhysicalDiskStatus defines the observed state of DiscoveredPhysicalDisk.
type DiscoveredPhysicalDiskStatus struct {
	// Available indicates whether the physical drive is present in the slot.
	// +optional
	Available *bool `json:"available,omitempty"`
	// Vendor is the disk manufacturer.
	// +optional
	Vendor *string `json:"vendor,omitempty"`
	// Model is the disk model name.
	// +optional
	Model *string `json:"model,omitempty"`
	// Serial is the disk serial number.
	// +optional
	Serial *string `json:"serial,omitempty"`
	// WWN is the World Wide Name of the disk.
	// +optional
	WWN *string `json:"wwn,omitempty"`
	// Size is the disk capacity in bytes.
	// +optional
	Size *int64 `json:"size,omitempty"`
	// Type is the disk media type.
	// +kubebuilder:validation:Enum=HDD;SSD;NVMe
	// +optional
	Type *string `json:"type,omitempty"`
	// JBOD indicates whether the disk is in JBOD (passthrough) mode.
	// +optional
	JBOD *bool `json:"jbod,omitempty"`
	// Status is the current disk status.
	// +kubebuilder:validation:Enum=Used;Available;Failed
	// +optional
	Status *string `json:"status,omitempty"`
	// Reason provides additional context for the current status.
	// +optional
	Reason *string `json:"reason,omitempty"`
	// DevicePath is the OS device path (e.g. "/dev/sda").
	// +optional
	DevicePath *string `json:"devicePath,omitempty"`
	// PermanentPath is the stable device path (e.g. "/dev/disk/by-id/wwn-0x...").
	// +optional
	PermanentPath *string `json:"permanentPath,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:selectablefield:JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="Available",type=boolean,JSONPath=`.status.available`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.type`
// +kubebuilder:printcolumn:name="Size",type=integer,JSONPath=`.status.size`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// DiscoveredPhysicalDisk is the Schema for the discoveredphysicaldisks API.
type DiscoveredPhysicalDisk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiscoveredPhysicalDiskSpec   `json:"spec,omitempty"`
	Status DiscoveredPhysicalDiskStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DiscoveredPhysicalDiskList contains a list of DiscoveredPhysicalDisk.
type DiscoveredPhysicalDiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveredPhysicalDisk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveredPhysicalDisk{}, &DiscoveredPhysicalDiskList{})
}
