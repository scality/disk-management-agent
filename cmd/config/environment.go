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
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

// ApplicationVersion is the version of the application.
// It is set at build time using ldflags.
//
//nolint:gochecknoglobals // This is set at build time.
var ApplicationVersion = "dev"

const ApplicationName = "disk-management-agent"

type Environment struct {
	NodeName          string `env:"NODE_NAME, required"`
	PodNamespace      string `env:"POD_NAMESPACE"`
	PodServiceAccount string `env:"POD_SERVICE_ACCOUNT"`
	StorcliPath       string `env:"STORCLI_PATH, default=/host/libexec/MegaRAID/storcli/storcli64"`
	PerccliPath       string `env:"PERCCLI_PATH, default=/host/libexec/MegaRAID/perccli/perccli64"`
	SsacliPath        string `env:"SSACLI_PATH, default=/host/libexec/ssacli"`
}

func NewEnvironment(ctx context.Context) (*Environment, error) {
	cfg := &Environment{}

	err := cfg.Load(ctx)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Environment) Load(ctx context.Context) error {
	err := envconfig.Process(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to process environment variables: %w", err)
	}

	return nil
}
