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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"disk-management-agent/pkg/domain"
)

// TestEnvironmentLoad checks how the environment is read: defaults apply to
// unset variables, set variables win, and NODE_NAME is required. go-envconfig
// v2 no longer runs decoders on unset variables, which must not change this.
func TestEnvironmentLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *Environment
		wantErr error
	}{
		{
			name: "defaults",
			env:  map[string]string{"NODE_NAME": "node-1"},
			want: &Environment{
				NodeName:     "node-1",
				StorcliPath:  "/host/libexec/MegaRAID/storcli/storcli64",
				PerccliPath:  "/host/libexec/MegaRAID/perccli/perccli64",
				Storcli2Path: "/host/libexec/MegaRAID/storcli2/storcli2",
				Perccli2Path: "/host/libexec/MegaRAID/perccli2/perccli2",
				SsacliPath:   "/host/libexec/ssacli",
			},
		},
		{
			name: "set variables win over defaults",
			env: map[string]string{
				"NODE_NAME": "node-1", "POD_NAMESPACE": "kube-system", "SSACLI_PATH": "/opt/ssacli",
			},
			want: &Environment{
				NodeName:     "node-1",
				PodNamespace: "kube-system",
				StorcliPath:  "/host/libexec/MegaRAID/storcli/storcli64",
				PerccliPath:  "/host/libexec/MegaRAID/perccli/perccli64",
				Storcli2Path: "/host/libexec/MegaRAID/storcli2/storcli2",
				Perccli2Path: "/host/libexec/MegaRAID/perccli2/perccli2",
				SsacliPath:   "/opt/ssacli",
			},
		},
		{
			name:    "NODE_NAME is required",
			env:     map[string]string{},
			wantErr: domain.ErrConfigLoad,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{
				"NODE_NAME", "POD_NAMESPACE", "POD_SERVICE_ACCOUNT", "STORCLI_PATH",
				"PERCCLI_PATH", "STORCLI2_PATH", "PERCCLI2_PATH", "SSACLI_PATH",
			} {
				// Unset, not empty: t.Setenv restores the variable after the
				// test, os.Unsetenv removes it for the test.
				t.Setenv(key, "")
				require.NoError(t, os.Unsetenv(key))
			}

			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			cfg, err := NewEnvironment(t.Context())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg)
		})
	}
}
