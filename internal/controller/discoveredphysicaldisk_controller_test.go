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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metalk8sv1alpha1 "disk-management-agent/api/v1alpha1"
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
			Size: 4000787030016,
			Type: "HDD",
		},
	}
}

var _ = Describe("DiscoveredPhysicalDisk Controller", func() {
	Context("When reconciling a resource", func() {
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

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DiscoveredPhysicalDiskReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				NodeName: "node-1",
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
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
