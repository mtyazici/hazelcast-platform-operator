package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hazelcastcomv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	mcconfig "github.com/hazelcast/hazelcast-platform-operator/test/e2e/config/managementcenter"
)

const (
	mcName = "managementcenter"
)

var _ = Describe("Management-Center", func() {

	var lookupKey = types.NamespacedName{
		Name:      mcName,
		Namespace: hzNamespace,
	}

	var controllerManagerName = types.NamespacedName{
		Name:      controllerManagerName(),
		Namespace: hzNamespace,
	}

	BeforeEach(func() {
		if !useExistingCluster() {
			Skip("End to end tests require k8s cluster. Set USE_EXISTING_CLUSTER=true")
		}

		if runningLocally() {
			return
		}
		By("Checking hazelcast-platform-controller-manager running", func() {
			controllerDep := &appsv1.Deployment{}
			Eventually(func() (int32, error) {
				return getDeploymentReadyReplicas(context.Background(), controllerManagerName, controllerDep)
			}, timeout, interval).Should(Equal(int32(1)))
		})

	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), emptyManagementCenter(), client.PropagationPolicy(v1.DeletePropagationForeground))).Should(Succeed())
		assertDoesNotExist(lookupKey, &hazelcastcomv1alpha1.ManagementCenter{})

		pvcLookupKey := types.NamespacedName{
			Name:      fmt.Sprintf("mancenter-storage-%s-0", lookupKey.Name),
			Namespace: lookupKey.Namespace,
		}
		deleteIfExists(pvcLookupKey, &corev1.PersistentVolumeClaim{})
	})

	create := func(mancenter *hazelcastcomv1alpha1.ManagementCenter) {
		By("Creating ManagementCenter CR", func() {
			Expect(k8sClient.Create(context.Background(), mancenter)).Should(Succeed())
		})

		By("Checking ManagementCenter CR running", func() {
			mc := &hazelcastcomv1alpha1.ManagementCenter{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, mc)
				Expect(err).ToNot(HaveOccurred())
				return isManagementCenterRunning(mc)
			}, timeout, interval).Should(BeTrue())
		})
	}

	createWithoutCheck := func(mancenter *hazelcastcomv1alpha1.ManagementCenter) {
		By("Creating ManagementCenter CR", func() {
			Expect(k8sClient.Create(context.Background(), mancenter)).Should(Succeed())
		})
	}

	Describe("Default ManagementCenter CR", func() {
		It("Should create ManagementCenter resources", func() {
			mc := mcconfig.Default(hzNamespace, ee)
			create(mc)

			By("Checking if it created PVC with correct size", func() {
				fetchedPVC := &corev1.PersistentVolumeClaim{}
				pvcLookupKey := types.NamespacedName{
					Name:      fmt.Sprintf("mancenter-storage-%s-0", lookupKey.Name),
					Namespace: lookupKey.Namespace,
				}
				assertExists(pvcLookupKey, fetchedPVC)

				expectedResourceList := corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				}
				Expect(fetchedPVC.Status.Capacity).Should(Equal(expectedResourceList))
			})

		})

	})

	Describe("ManagementCenter CR without Persistence", func() {
		It("Should create ManagementCenter resources and no PVC", func() {
			mc := mcconfig.PersistenceDisabled(hzNamespace, ee)
			create(mc)

			By("Checking if PVC does not exist", func() {
				fetchedPVC := &corev1.PersistentVolumeClaim{}
				pvcLookupKey := types.NamespacedName{
					Name:      fmt.Sprintf("mancenter-storage-%s-0", lookupKey.Name),
					Namespace: lookupKey.Namespace,
				}
				assertDoesNotExist(pvcLookupKey, fetchedPVC)
			})

		})

	})

	Describe("External API errors", func() {
		assertStatusEventually := func(phase hazelcastcomv1alpha1.Phase) {
			mc := &hazelcastcomv1alpha1.ManagementCenter{}
			Eventually(func() hazelcastcomv1alpha1.Phase {
				err := k8sClient.Get(context.Background(), lookupKey, mc)
				Expect(err).ToNot(HaveOccurred())
				return mc.Status.Phase
			}, timeout, interval).Should(Equal(phase))
			Expect(mc.Status.Message).Should(Not(BeEmpty()))
		}

		It("should be reflected to Management CR status", func() {
			createWithoutCheck(mcconfig.Faulty(hzNamespace, ee))
			assertStatusEventually(hazelcastcomv1alpha1.Failed)

		})
	})

})

func emptyManagementCenter() *hazelcastcomv1alpha1.ManagementCenter {
	return &hazelcastcomv1alpha1.ManagementCenter{
		ObjectMeta: v1.ObjectMeta{
			Name:      mcName,
			Namespace: hzNamespace,
		},
	}
}

func isManagementCenterRunning(mc *hazelcastcomv1alpha1.ManagementCenter) bool {
	return mc.Status.Phase == "Running"
}
