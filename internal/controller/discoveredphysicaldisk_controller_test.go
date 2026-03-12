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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/scality/raidmgmt/pkg/domain/entities/physicaldrive"
	"github.com/scality/raidmgmt/pkg/domain/entities/raidcontroller"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
	"disk-management-agent/pkg/domain"
	"disk-management-agent/pkg/infrastructure/discovereddrivecache"
	"disk-management-agent/pkg/usecase"
)

// newDisk is a test helper that creates a DiscoveredPhysicalDisk resource.
//
//nolint:unparam // namespace is intentionally parameterized for reuse across test contexts.
func newDisk(name, namespace, nodeName string) *metalk8sv1alpha1.DiscoveredPhysicalDisk {
	return &metalk8sv1alpha1.DiscoveredPhysicalDisk{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: metalk8sv1alpha1.DiscoveredPhysicalDiskSpec{
			NodeName: nodeName,
			Controller: metalk8sv1alpha1.ControllerRef{
				Type: "MegaRAID",
				ID:   0,
			},
			ID: "0:1:2",
			Slot: metalk8sv1alpha1.SlotLocation{
				Port:      "0",
				Enclosure: "1",
				Bay:       "2",
			},
		},
	}
}

var _ = Describe("DiscoveredPhysicalDisk Controller", func() {
	Context("When reconciling a resource with drive data in the cache", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		discoveredphysicaldisk := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind DiscoveredPhysicalDisk")
			err := k8sClient.Get(ctx, typeNamespacedName, discoveredphysicaldisk)
			if err != nil && errors.IsNotFound(err) {
				resource := newDisk(resourceName, "default", "node-1")
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DiscoveredPhysicalDisk")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should fill the status fields from the cached drive data", func() {
			By("populating the drive cache with test data")
			cache := discovereddrivecache.NewInMemory()
			cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{
				resourceName: {
					ControllerType: "MegaRAID",
					ControllerID:   0,
					PhysicalDrive: &physicaldrive.PhysicalDrive{
						Metadata: &physicaldrive.Metadata{
							CtrlMetadata: &raidcontroller.Metadata{ID: 0},
							ID:           "0:1:2",
						},
						Vendor: "Seagate",
						Model:  "ST4000NM0033",
						Serial: "Z1Z2Z3Z4",
						WWN:    "5000C50012345678",
						Size:   4000787030016,
						Type:   physicaldrive.DiskTypeHDD,
						Status: physicaldrive.PDStatusUsed,
						JBOD:   true,
						Reason: "part of logical volume",
					},
				},
			})

			By("Reconciling the created resource")
			controllerReconciler := &DiscoveredPhysicalDiskReconciler{
				Client:           k8sClient,
				Scheme:           k8sClient.Scheme(),
				NodeName:         "node-1",
				ReconcileUseCase: usecase.NewReconcileDiscoveredPhysicalDisk(cache),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the status was updated")
			updated := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())

			Expect(*updated.Status.Available).To(BeTrue())
			Expect(*updated.Status.Vendor).To(Equal("Seagate"))
			Expect(*updated.Status.Model).To(Equal("ST4000NM0033"))
			Expect(*updated.Status.Serial).To(Equal("Z1Z2Z3Z4"))
			Expect(*updated.Status.WWN).To(Equal("5000C50012345678"))
			Expect(*updated.Status.Size).To(Equal(int64(4000787030016)))
			Expect(*updated.Status.Type).To(Equal("HDD"))
			Expect(*updated.Status.JBOD).To(BeTrue())
			Expect(*updated.Status.Status).To(Equal("Used"))
			Expect(*updated.Status.Reason).To(Equal("part of logical volume"))
		})
	})

	Context("When reconciling with an empty cache (cache ready, drive not found)", func() {
		const resourceName = "test-resource-unavailable"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			resource := newDisk(resourceName, "default", "node-1")
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should set available to false", func() {
			cache := discovereddrivecache.NewInMemory()
			cache.Replace(map[string]*domain.DiscoveredPhysicalDrive{})

			controllerReconciler := &DiscoveredPhysicalDiskReconciler{
				Client:           k8sClient,
				Scheme:           k8sClient.Scheme(),
				NodeName:         "node-1",
				ReconcileUseCase: usecase.NewReconcileDiscoveredPhysicalDisk(cache),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			updated := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())

			Expect(*updated.Status.Available).To(BeFalse())
		})
	})

	Context("When reconciling before the cache is ready", func() {
		const resourceName = "test-resource-cache-pending"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			resource := newDisk(resourceName, "default", "node-1")
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should requeue after a delay", func() {
			cache := discovereddrivecache.NewInMemory()

			controllerReconciler := &DiscoveredPhysicalDiskReconciler{
				Client:           k8sClient,
				Scheme:           k8sClient.Scheme(),
				NodeName:         "node-1",
				ReconcileUseCase: usecase.NewReconcileDiscoveredPhysicalDisk(cache),
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(cacheNotReadyRequeueDelay))
		})
	})

	Context("When using field selectors on spec.nodeName", func() {
		ctx := context.Background()

		BeforeEach(func() {
			By("creating disks on different nodes")
			Expect(k8sClient.Create(ctx, newDisk("disk-node1-a", "default", "node-1"))).To(Succeed())
			Expect(k8sClient.Create(ctx, newDisk("disk-node1-b", "default", "node-1"))).To(Succeed())
			Expect(k8sClient.Create(ctx, newDisk("disk-node2-a", "default", "node-2"))).To(Succeed())
		})

		AfterEach(func() {
			By("cleaning up all test disks")
			for _, name := range []string{"disk-node1-a", "disk-node1-b", "disk-node2-a"} {
				resource := &metalk8sv1alpha1.DiscoveredPhysicalDisk{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: "default"}, resource)
				if err == nil {
					Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
				}
			}
		})

		It("should only return disks matching the selected node when using a field selector", func() {
			By("listing disks with field selector spec.nodeName=node-1")
			node1Disks := &metalk8sv1alpha1.DiscoveredPhysicalDiskList{}
			Expect(k8sClient.List(ctx, node1Disks,
				client.InNamespace("default"),
				client.MatchingFields{"spec.nodeName": "node-1"},
			)).To(Succeed())
			Expect(node1Disks.Items).To(HaveLen(2))
			for _, disk := range node1Disks.Items {
				Expect(disk.Spec.NodeName).To(Equal("node-1"))
			}

			By("listing disks with field selector spec.nodeName=node-2")
			node2Disks := &metalk8sv1alpha1.DiscoveredPhysicalDiskList{}
			Expect(k8sClient.List(ctx, node2Disks,
				client.InNamespace("default"),
				client.MatchingFields{"spec.nodeName": "node-2"},
			)).To(Succeed())
			Expect(node2Disks.Items).To(HaveLen(1))
			Expect(node2Disks.Items[0].Spec.NodeName).To(Equal("node-2"))

			By("listing disks with field selector for a non-existent node")
			noDisks := &metalk8sv1alpha1.DiscoveredPhysicalDiskList{}
			Expect(k8sClient.List(ctx, noDisks,
				client.InNamespace("default"),
				client.MatchingFields{"spec.nodeName": "node-999"},
			)).To(Succeed())
			Expect(noDisks.Items).To(BeEmpty())
		})
	})
})
