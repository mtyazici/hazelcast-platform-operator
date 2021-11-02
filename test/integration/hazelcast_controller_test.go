package integration

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/controllers/naming"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hazelcast controller", func() {
	const (
		hzKeyName = "hazelcast-test"
		namespace = "default"
		finalizer = n.Finalizer

		timeout  = time.Second * 10
		interval = time.Millisecond * 250

		clusterSize      = n.DefaultClusterSize
		repository       = n.HazelcastRepo
		version          = n.HazelcastVersion
		licenseKeySecret = n.LicenseKeySecret
	)

	lookupKey := types.NamespacedName{
		Name:      hzKeyName,
		Namespace: namespace,
	}

	labelFilter := client.MatchingLabels{
		n.ApplicationNameLabel:      n.Hazelcast,
		n.ApplicationManagedByLabel: n.OperatorName,
	}

	Create := func(hz *hazelcastv1alpha1.Hazelcast) {
		By("creating the CR with specs successfully")
		Expect(k8sClient.Create(context.Background(), hz)).Should(Succeed())
		time.Sleep(time.Second * 2)
	}

	Update := func(hz *hazelcastv1alpha1.Hazelcast) {
		By("updating the CR with specs successfully")
		Expect(k8sClient.Update(context.Background(), hz)).Should(Succeed())
		time.Sleep(time.Second * 2)
	}

	Fetch := func() *hazelcastv1alpha1.Hazelcast {
		By("fetching Hazelcast")
		fetchedCR := &hazelcastv1alpha1.Hazelcast{}
		Eventually(func() bool {
			err := k8sClient.Get(context.Background(), lookupKey, fetchedCR)
			if err != nil {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
		return fetchedCR
	}

	Delete := func() {
		By("expecting to delete CR successfully")
		Eventually(func() error {
			fetchedCR := &hazelcastv1alpha1.Hazelcast{}
			_ = k8sClient.Get(context.Background(), lookupKey, fetchedCR)
			return k8sClient.Delete(context.Background(), fetchedCR)
		}, timeout, interval).Should(Succeed())

		By("expecting to CR delete finish")
		Eventually(func() error {
			return k8sClient.Get(context.Background(), lookupKey, &hazelcastv1alpha1.Hazelcast{})
		}, timeout, interval).ShouldNot(Succeed())
	}

	EnsureStatus := func(hz *hazelcastv1alpha1.Hazelcast) {
		By("ensuring that the status is correct")
		Expect(hz.Status.Phase).Should(Equal(hazelcastv1alpha1.Pending))
	}

	EnsureFailedStatus := func(hz *hazelcastv1alpha1.Hazelcast) {
		By("ensuring that the status is failed")
		Expect(hz.Status.Phase).Should(Equal(hazelcastv1alpha1.Failed))
	}

	Context("Hazelcast CustomResource with default specs", func() {
		It("should handle CR and sub resources correctly", func() {
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      clusterSize,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
				},
			}
			Create(hz)

			fetchedCR := Fetch()
			Expect(fetchedCR.Spec.ClusterSize).Should(Equal(int32(clusterSize)))
			Expect(fetchedCR.Spec.Repository).Should(Equal(repository))
			Expect(fetchedCR.Spec.Version).Should(Equal(version))
			Expect(fetchedCR.Spec.LicenseKeySecret).Should(Equal(licenseKeySecret))
			EnsureStatus(fetchedCR)

			By("ensuring the finalizer added successfully")
			Expect(fetchedCR.Finalizers).To(ContainElement(finalizer))

			By("creating the sub resources successfully")
			expectedOwnerReference := metav1.OwnerReference{
				Kind:               "Hazelcast",
				APIVersion:         "hazelcast.com/v1alpha1",
				UID:                fetchedCR.UID,
				Name:               fetchedCR.Name,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			}

			fetchedClusterRole := &rbacv1.ClusterRole{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, fetchedClusterRole)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			fetchedServiceAccount := &corev1.ServiceAccount{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, fetchedServiceAccount)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(fetchedServiceAccount.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			fetchedClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, fetchedClusterRoleBinding)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			fetchedService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, fetchedService)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(fetchedService.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			fetchedSts := &v1.StatefulSet{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), lookupKey, fetchedSts)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(fetchedSts.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
			Expect(fetchedSts.Spec.Template.Spec.Containers[0].Image).Should(Equal(fetchedCR.DockerImage()))

			Delete()

			By("Expecting to ClusterRole removed via finalizer")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), lookupKey, &rbacv1.ClusterRole{})
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Hazelcast CustomResource with expose externally", func() {
		FetchServices := func(waitForN int) *corev1.ServiceList {
			serviceList := &corev1.ServiceList{}
			Eventually(func() bool {
				err := k8sClient.List(context.Background(), serviceList, client.InNamespace(namespace), labelFilter)
				if err != nil || len(serviceList.Items) != waitForN {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			return serviceList
		}

		It("should create Hazelcast cluster exposed for unisocket client", func() {
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      clusterSize,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
					ExposeExternally: hazelcastv1alpha1.ExposeExternallyConfiguration{
						Type:                 hazelcastv1alpha1.ExposeExternallyTypeUnisocket,
						DiscoveryServiceType: corev1.ServiceTypeNodePort,
					},
				},
			}
			Create(hz)

			fetchedCR := Fetch()
			Expect(fetchedCR.Spec.ExposeExternally.Type).Should(Equal(hazelcastv1alpha1.ExposeExternallyTypeUnisocket))
			Expect(fetchedCR.Spec.ExposeExternally.DiscoveryServiceType).Should(Equal(corev1.ServiceTypeNodePort))
			EnsureStatus(fetchedCR)

			By("checking created services")
			serviceList := FetchServices(1)

			service := serviceList.Items[0]
			Expect(service.Name).Should(Equal(hz.Name))
			Expect(service.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))

			Delete()
		})

		It("should create Hazelcast cluster exposed for smart client", func() {
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      clusterSize,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
					ExposeExternally: hazelcastv1alpha1.ExposeExternallyConfiguration{
						Type:                 hazelcastv1alpha1.ExposeExternallyTypeSmart,
						DiscoveryServiceType: corev1.ServiceTypeNodePort,
						MemberAccess:         hazelcastv1alpha1.MemberAccessNodePortExternalIP,
					},
				},
			}
			Create(hz)

			fetchedCR := Fetch()
			Expect(fetchedCR.Spec.ExposeExternally.Type).Should(Equal(hazelcastv1alpha1.ExposeExternallyTypeSmart))
			Expect(fetchedCR.Spec.ExposeExternally.DiscoveryServiceType).Should(Equal(corev1.ServiceTypeNodePort))
			Expect(fetchedCR.Spec.ExposeExternally.MemberAccess).Should(Equal(hazelcastv1alpha1.MemberAccessNodePortExternalIP))
			EnsureStatus(fetchedCR)

			By("checking created services")
			serviceList := FetchServices(4)

			for _, s := range serviceList.Items {
				if s.Name == lookupKey.Name {
					// discovery service
					Expect(s.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
				} else {
					// member access service
					Expect(s.Name).Should(ContainSubstring(lookupKey.Name))
					Expect(s.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
				}
			}

			Delete()
		})

		It("should scale Hazelcast cluster exposed for smart client", func() {
			By("creating the cluster of size 3")
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      3,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
					ExposeExternally: hazelcastv1alpha1.ExposeExternallyConfiguration{
						Type:                 hazelcastv1alpha1.ExposeExternallyTypeSmart,
						DiscoveryServiceType: corev1.ServiceTypeNodePort,
						MemberAccess:         hazelcastv1alpha1.MemberAccessNodePortExternalIP,
					},
				},
			}
			Create(hz)
			fetchedCR := Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(4)

			By("scaling the cluster to 6 members")
			fetchedCR.Spec.ClusterSize = 6
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(7)

			By("scaling the cluster to 1 member")
			fetchedCR.Spec.ClusterSize = 1
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(2)

			By("deleting the cluster")
			Delete()
		})

		It("should allow updating expose externally configuration", func() {
			By("creating the cluster with smart client")
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      3,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
					ExposeExternally: hazelcastv1alpha1.ExposeExternallyConfiguration{
						Type:                 hazelcastv1alpha1.ExposeExternallyTypeSmart,
						DiscoveryServiceType: corev1.ServiceTypeNodePort,
						MemberAccess:         hazelcastv1alpha1.MemberAccessNodePortExternalIP,
					},
				},
			}
			Create(hz)
			fetchedCR := Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(4)

			By("updating type to unisocket")
			fetchedCR.Spec.ExposeExternally.Type = hazelcastv1alpha1.ExposeExternallyTypeUnisocket
			fetchedCR.Spec.ExposeExternally.MemberAccess = ""
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(1)

			By("updating discovery service to LoadBalancer")
			fetchedCR.Spec.ExposeExternally.DiscoveryServiceType = corev1.ServiceTypeLoadBalancer
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			serviceList := FetchServices(1)
			Expect(serviceList.Items[0].Spec.Type).Should(Equal(corev1.ServiceTypeLoadBalancer))

			By("updating type to smart")
			fetchedCR.Spec.ExposeExternally.Type = hazelcastv1alpha1.ExposeExternallyTypeSmart
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			FetchServices(4)

			By("deleting expose externally configuration")
			fetchedCR.Spec.ExposeExternally = hazelcastv1alpha1.ExposeExternallyConfiguration{}
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			serviceList = FetchServices(1)
			Expect(serviceList.Items[0].Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))

			Delete()
		})

		It("should return expected messages when exposeExternally is misconfigured", func() {
			By("creating the cluster with unisocket client with incorrect configuration")
			hz := &hazelcastv1alpha1.Hazelcast{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lookupKey.Name,
					Namespace: lookupKey.Namespace,
				},
				Spec: hazelcastv1alpha1.HazelcastSpec{
					ClusterSize:      3,
					Repository:       repository,
					Version:          version,
					LicenseKeySecret: licenseKeySecret,
					ExposeExternally: hazelcastv1alpha1.ExposeExternallyConfiguration{
						Type:                 hazelcastv1alpha1.ExposeExternallyTypeUnisocket,
						DiscoveryServiceType: corev1.ServiceTypeNodePort,
						MemberAccess:         hazelcastv1alpha1.MemberAccessLoadBalancer,
					},
				},
			}
			Create(hz)
			fetchedCR := Fetch()
			EnsureFailedStatus(fetchedCR)
			Expect(fetchedCR.Status.Message).To(Equal("error validating new Spec: when exposeExternally.type is set to \"Unisocket\", exposeExternally.memberAccess must not be set"))

			By("fixing the incorrect configuration")
			fetchedCR.Spec.ExposeExternally.MemberAccess = ""
			Update(fetchedCR)
			fetchedCR = Fetch()
			EnsureStatus(fetchedCR)
			Expect(fetchedCR.Status.Message).To(BeEmpty())

			Delete()
		})
	})
})