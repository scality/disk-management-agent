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

package discoveredphysicaldiskstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
)

var testScheme *runtime.Scheme

func init() {
	testScheme = runtime.NewScheme()
	_ = metalk8sv1alpha1.AddToScheme(testScheme)
}

//nolint:unparam // namespace is intentionally parameterized for reuse across test contexts.
func newTestDisk(name, namespace string) *metalk8sv1alpha1.DiscoveredPhysicalDisk {
	return &metalk8sv1alpha1.DiscoveredPhysicalDisk{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: metalk8sv1alpha1.DiscoveredPhysicalDiskSpec{
			NodeName: "node-1",
			Controller: metalk8sv1alpha1.ControllerRef{
				Type: "MegaRAID",
				ID:   0,
			},
			ID:   "0:1:0",
			Slot: metalk8sv1alpha1.SlotLocation{Port: "0", Enclosure: "1", Bay: "0"},
		},
	}
}

func TestGet_Exists(t *testing.T) {
	seed := newTestDisk("disk-1", "default")
	k8s := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(seed).Build()
	store := NewKubernetes(k8s)

	result, err := store.Get(context.Background(), "default", "disk-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "disk-1", result.Name)
	assert.Equal(t, "node-1", result.Spec.NodeName)
}

func TestGet_NotFound(t *testing.T) {
	k8s := fake.NewClientBuilder().WithScheme(testScheme).Build()
	store := NewKubernetes(k8s)

	result, err := store.Get(context.Background(), "default", "nonexistent")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCreate_Success(t *testing.T) {
	k8s := fake.NewClientBuilder().WithScheme(testScheme).Build()
	store := NewKubernetes(k8s)
	disk := newTestDisk("disk-new", "default")

	err := store.Create(context.Background(), disk)
	require.NoError(t, err)

	got, err := store.Get(context.Background(), "default", "disk-new")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "disk-new", got.Name)
	assert.Equal(t, "MegaRAID", got.Spec.Controller.Type)
}

func TestCreate_AlreadyExists(t *testing.T) {
	seed := newTestDisk("disk-dup", "default")
	k8s := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(seed).Build()
	store := NewKubernetes(k8s)

	dup := newTestDisk("disk-dup", "default")
	err := store.Create(context.Background(), dup)

	require.Error(t, err)
	assert.True(t, apierrors.IsAlreadyExists(err))
}
