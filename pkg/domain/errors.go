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

import "github.com/scality/go-errors"

//nolint:gochecknoglobals // Sentinel errors are package-level by design.
var (
	ErrControllerDiscovery = errors.New("controller discovery failed")
	ErrDriveDiscovery      = errors.New("drive discovery failed")
	ErrVolumeDiscovery     = errors.New("volume discovery failed")
	ErrDiskStoreOperation  = errors.New("disk store operation failed")
	ErrConfigLoad          = errors.New("configuration load failed")
	ErrReconciliation      = errors.New("reconciliation failed")
	ErrValidation          = errors.New("validation failed")
)
