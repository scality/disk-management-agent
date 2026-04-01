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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeCRName(t *testing.T) {
	tests := []struct {
		name           string
		nodeName       string
		controllerType string
		controllerID   int
		slotID         string
		expected       string
	}{
		{
			name:           "MegaRAID with port:enclosure:bay slot",
			nodeName:       "node-1",
			controllerType: "MegaRAID",
			controllerID:   0,
			slotID:         "0:1:2",
			expected:       "node-1-megaraid-0-0-1-2",
		},
		{
			name:           "SmartArray with mixed-case slot",
			nodeName:       "worker-3",
			controllerType: "SmartArray",
			controllerID:   1,
			slotID:         "1I:1:4",
			expected:       "worker-3-smartarray-1-1i-1-4",
		},
		{
			name:           "bay-only slot",
			nodeName:       "node-1",
			controllerType: "MegaRAID",
			controllerID:   0,
			slotID:         "5",
			expected:       "node-1-megaraid-0-5",
		},
		{
			name:           "enclosure:bay slot",
			nodeName:       "node-2",
			controllerType: "MegaRAID",
			controllerID:   0,
			slotID:         "252:0",
			expected:       "node-2-megaraid-0-252-0",
		},
		{
			name:           "no collision between different slot hierarchies",
			nodeName:       "node-1",
			controllerType: "MegaRAID",
			controllerID:   0,
			slotID:         "25:2:0",
			expected:       "node-1-megaraid-0-25-2-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeCRName(tt.nodeName, tt.controllerType, tt.controllerID, tt.slotID)
			assert.Equal(t, tt.expected, got)
		})
	}
}
