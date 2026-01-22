/*
Copyright 2026 CraightonH.

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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
	"github.com/CraightonH/boilerr/internal/resources"
)

// createTestGameDefinition creates a GameDefinition for testing and waits for it to be ready.
func createTestGameDefinition(name string, appID int32) *boilerrv1alpha1.GameDefinition {
	gameDef := &boilerrv1alpha1.GameDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: boilerrv1alpha1.GameDefinitionSpec{
			AppId:   appID,
			Command: "/serverfiles/start.sh",
			Ports: []boilerrv1alpha1.ServerPort{
				{Name: "game", ContainerPort: 27015, Protocol: corev1.ProtocolUDP},
			},
		},
	}
	ExpectWithOffset(1, k8sClient.Create(ctx, gameDef)).Should(Succeed())

	// Wait for the GameDefinition to be marked ready
	gameDefKey := types.NamespacedName{Name: name}
	EventuallyWithOffset(1, func() bool {
		gd := &boilerrv1alpha1.GameDefinition{}
		if err := k8sClient.Get(ctx, gameDefKey, gd); err != nil {
			return false
		}
		return gd.Status.Ready
	}, time.Second*10, time.Millisecond*250).Should(BeTrue())

	return gameDef
}

// deleteTestGameDefinition deletes a GameDefinition if it exists.
func deleteTestGameDefinition(name string) {
	gameDef := &boilerrv1alpha1.GameDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_ = k8sClient.Delete(ctx, gameDef)
	EventuallyWithOffset(1, func() bool {
		err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, &boilerrv1alpha1.GameDefinition{})
		return errors.IsNotFound(err)
	}, time.Second*10, time.Millisecond*250).Should(BeTrue())
}

var _ = Describe("SteamServer Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a SteamServer", func() {
		It("Should create child resources", func() {
			By("Creating the GameDefinition first")
			gameDef := createTestGameDefinition("valheim", 896660)
			defer deleteTestGameDefinition("valheim")

			By("Creating a new SteamServer")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-valheim",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: gameDef.Name,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 2456, Protocol: corev1.ProtocolUDP},
						{Name: "query", ContainerPort: 2457, Protocol: corev1.ProtocolUDP},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("20Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, steamServer) }()

			steamServerLookupKey := types.NamespacedName{Name: "test-valheim", Namespace: "default"}
			createdSteamServer := &boilerrv1alpha1.SteamServer{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, steamServerLookupKey, createdSteamServer)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Checking that the StatefulSet is created")
			stsLookupKey := types.NamespacedName{Name: "test-valheim", Namespace: "default"}
			createdSts := &appsv1.StatefulSet{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, stsLookupKey, createdSts)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSts.Name).Should(Equal("test-valheim"))
			Expect(*createdSts.Spec.Replicas).Should(Equal(int32(1)))
			Expect(createdSts.Spec.Template.Spec.Containers).Should(HaveLen(1))
			Expect(createdSts.Spec.Template.Spec.InitContainers).Should(HaveLen(1))

			By("Checking that the Service is created")
			svcLookupKey := types.NamespacedName{Name: "test-valheim", Namespace: "default"}
			createdSvc := &corev1.Service{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, svcLookupKey, createdSvc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSvc.Name).Should(Equal("test-valheim"))
			Expect(createdSvc.Spec.Type).Should(Equal(corev1.ServiceTypeLoadBalancer))
			Expect(createdSvc.Spec.Ports).Should(HaveLen(2))

			By("Checking that the PVC is created")
			pvcLookupKey := types.NamespacedName{Name: "test-valheim-data", Namespace: "default"}
			createdPvc := &corev1.PersistentVolumeClaim{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, pvcLookupKey, createdPvc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdPvc.Name).Should(Equal("test-valheim-data"))
			expectedSize := resource.MustParse("20Gi")
			Expect(createdPvc.Spec.Resources.Requests[corev1.ResourceStorage]).Should(Equal(expectedSize))

			By("Checking that the finalizer is added")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, steamServerLookupKey, createdSteamServer)
				if err != nil {
					return false
				}
				for _, f := range createdSteamServer.Finalizers {
					if f == FinalizerName {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When creating a SteamServer with config files", func() {
		It("Should create a ConfigMap", func() {
			By("Creating the GameDefinition first")
			createTestGameDefinition("test-game-config", 123456)
			defer deleteTestGameDefinition("test-game-config")

			By("Creating a SteamServer with config files")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-with-config",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: "test-game-config",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
					ConfigFiles: []boilerrv1alpha1.ConfigFile{
						{
							Path:    "/config/server.cfg",
							Content: "hostname MyServer",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, steamServer) }()

			By("Checking that the ConfigMap is created")
			cmLookupKey := types.NamespacedName{
				Name:      resources.ConfigMapName("test-with-config"),
				Namespace: "default",
			}
			createdCm := &corev1.ConfigMap{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, cmLookupKey, createdCm)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdCm.Data).Should(HaveKey("config-0"))
			Expect(createdCm.Data["config-0"]).Should(Equal("hostname MyServer"))
		})
	})

	Context("When creating a SteamServer without config files", func() {
		It("Should not create a ConfigMap", func() {
			By("Creating the GameDefinition first")
			createTestGameDefinition("test-game-noconfig", 123456)
			defer deleteTestGameDefinition("test-game-noconfig")

			By("Creating a SteamServer without config files")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-no-config",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: "test-game-noconfig",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, steamServer) }()

			By("Waiting for StatefulSet to be created")
			stsLookupKey := types.NamespacedName{Name: "test-no-config", Namespace: "default"}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, stsLookupKey, &appsv1.StatefulSet{})
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Checking that ConfigMap is not created")
			cmLookupKey := types.NamespacedName{
				Name:      resources.ConfigMapName("test-no-config"),
				Namespace: "default",
			}
			cm := &corev1.ConfigMap{}
			Consistently(func() bool {
				err := k8sClient.Get(ctx, cmLookupKey, cm)
				return errors.IsNotFound(err)
			}, time.Second*2, interval).Should(BeTrue())
		})
	})

	Context("When creating a SteamServer with NodePort service type", func() {
		It("Should create a NodePort Service", func() {
			By("Creating the GameDefinition first")
			createTestGameDefinition("test-game-nodeport", 123456)
			defer deleteTestGameDefinition("test-game-nodeport")

			By("Creating a SteamServer with NodePort")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-nodeport",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: "test-game-nodeport",
					ServiceType:    corev1.ServiceTypeNodePort,
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, steamServer) }()

			By("Checking the Service type")
			svcLookupKey := types.NamespacedName{Name: "test-nodeport", Namespace: "default"}
			createdSvc := &corev1.Service{}

			Eventually(func() corev1.ServiceType {
				err := k8sClient.Get(ctx, svcLookupKey, createdSvc)
				if err != nil {
					return ""
				}
				return createdSvc.Spec.Type
			}, timeout, interval).Should(Equal(corev1.ServiceTypeNodePort))
		})
	})

	Context("When deleting a SteamServer", func() {
		It("Should remove the finalizer and allow deletion", func() {
			By("Creating the GameDefinition first")
			createTestGameDefinition("test-game-deletion", 123456)
			defer deleteTestGameDefinition("test-game-deletion")

			By("Creating a SteamServer")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deletion",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: "test-game-deletion",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())

			steamServerLookupKey := types.NamespacedName{Name: "test-deletion", Namespace: "default"}

			By("Waiting for the finalizer to be added")
			Eventually(func() bool {
				ss := &boilerrv1alpha1.SteamServer{}
				err := k8sClient.Get(ctx, steamServerLookupKey, ss)
				if err != nil {
					return false
				}
				for _, f := range ss.Finalizers {
					if f == FinalizerName {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SteamServer")
			Expect(k8sClient.Delete(ctx, steamServer)).Should(Succeed())

			By("Verifying the SteamServer is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, steamServerLookupKey, &boilerrv1alpha1.SteamServer{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When updating a SteamServer", func() {
		It("Should update child resources", func() {
			By("Creating the GameDefinition first")
			createTestGameDefinition("test-game-update", 123456)
			defer deleteTestGameDefinition("test-game-update")

			By("Creating a SteamServer")
			steamServer := &boilerrv1alpha1.SteamServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-update",
					Namespace: "default",
				},
				Spec: boilerrv1alpha1.SteamServerSpec{
					GameDefinition: "test-game-update",
					Ports: []boilerrv1alpha1.ServerPort{
						{Name: "game", ContainerPort: 27015},
					},
					Storage: &boilerrv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, steamServer)).Should(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, steamServer) }()

			steamServerLookupKey := types.NamespacedName{Name: "test-update", Namespace: "default"}

			By("Waiting for initial resources to be created")
			svcLookupKey := types.NamespacedName{Name: "test-update", Namespace: "default"}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, svcLookupKey, &corev1.Service{})
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Updating the SteamServer to add another port")
			updatedServer := &boilerrv1alpha1.SteamServer{}
			Expect(k8sClient.Get(ctx, steamServerLookupKey, updatedServer)).Should(Succeed())
			updatedServer.Spec.Ports = append(updatedServer.Spec.Ports, boilerrv1alpha1.ServerPort{
				Name:          "rcon",
				ContainerPort: 27016,
				Protocol:      corev1.ProtocolTCP,
			})
			Expect(k8sClient.Update(ctx, updatedServer)).Should(Succeed())

			By("Verifying the Service is updated with the new port")
			Eventually(func() int {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, svcLookupKey, svc)
				if err != nil {
					return 0
				}
				return len(svc.Spec.Ports)
			}, timeout, interval).Should(Equal(2))
		})
	})
})

var _ = Describe("Controller Helper Functions", func() {
	Context("commonLabels", func() {
		It("Should return the correct labels", func() {
			labels := commonLabels("my-server", "valheim")
			Expect(labels["app.kubernetes.io/name"]).To(Equal("steamserver"))
			Expect(labels["app.kubernetes.io/instance"]).To(Equal("my-server"))
			Expect(labels["app.kubernetes.io/managed-by"]).To(Equal("boilerr"))
			Expect(labels["boilerr.dev/game"]).To(Equal("valheim"))
		})

		It("Should not include game label when empty", func() {
			labels := commonLabels("my-server", "")
			Expect(labels).NotTo(HaveKey("boilerr.dev/game"))
		})
	})

	Context("hasLabels", func() {
		It("Should return true when all labels match", func() {
			actual := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			Expect(hasLabels(actual, expected)).To(BeTrue())
		})

		It("Should return false when labels are missing", func() {
			actual := map[string]string{
				"key1": "value1",
			}
			expected := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			Expect(hasLabels(actual, expected)).To(BeFalse())
		})

		It("Should return false when label values differ", func() {
			actual := map[string]string{
				"key1": "different",
			}
			expected := map[string]string{
				"key1": "value1",
			}
			Expect(hasLabels(actual, expected)).To(BeFalse())
		})
	})

	Context("portsEqual", func() {
		It("Should return true for equal ports", func() {
			a := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			b := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			Expect(portsEqual(a, b)).To(BeTrue())
		})

		It("Should return false for different lengths", func() {
			a := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			b := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
				{Name: "query", Port: 27016, Protocol: corev1.ProtocolUDP},
			}
			Expect(portsEqual(a, b)).To(BeFalse())
		})

		It("Should return false for different names", func() {
			a := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			b := []boilerrv1alpha1.PortStatus{
				{Name: "other", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			Expect(portsEqual(a, b)).To(BeFalse())
		})

		It("Should return false for different ports", func() {
			a := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			b := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27016, Protocol: corev1.ProtocolUDP},
			}
			Expect(portsEqual(a, b)).To(BeFalse())
		})

		It("Should return false for different protocols", func() {
			a := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolUDP},
			}
			b := []boilerrv1alpha1.PortStatus{
				{Name: "game", Port: 27015, Protocol: corev1.ProtocolTCP},
			}
			Expect(portsEqual(a, b)).To(BeFalse())
		})

		It("Should return true for empty slices", func() {
			var a, b []boilerrv1alpha1.PortStatus
			Expect(portsEqual(a, b)).To(BeTrue())
		})
	})
})
