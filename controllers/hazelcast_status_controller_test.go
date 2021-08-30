package controllers

import (
	"context"
	"github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	hzTypes "github.com/hazelcast/hazelcast-go-client/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"time"
)

var _ = Describe("Hazelcast status", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	var hzClient = HazelcastClient{
		MemberMap:           make(map[string]bool),
		memberEventsChannel: make(chan event.GenericEvent),
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "hazelcast",
		},
	}

	Context("Hazelcast membership listener", func() {
		It("Should add new member to the map", func() {
			stateChanged := cluster.MembershipStateChanged{
				Member: cluster.MemberInfo{
					Address: cluster.NewAddress("127.0.0.1", 5701),
					UUID:    hzTypes.NewUUID(),
				},
				State: cluster.MembershipStateAdded,
			}
			go getStatusUpdateListener(hzClient)(stateChanged)

			Eventually(func() bool {
				return hzClient.MemberMap[stateChanged.Member.String()]
			}, timeout, interval).Should(BeTrue())

			Eventually(func() event.GenericEvent {
				select {
				case e := <-hzClient.memberEventsChannel:
					return e
				default:
					return event.GenericEvent{}
				}
			}).Should(Equal(event.GenericEvent{
				Object: &v1alpha1.Hazelcast{ObjectMeta: metav1.ObjectMeta{
					Namespace: hzClient.NamespacedName.Namespace,
					Name:      hzClient.NamespacedName.Name,
				}}}))
		})

		It("Should remove the existing member for the map", func() {
			existingMember := cluster.MemberInfo{
				Address: cluster.NewAddress("172.10.0.1", 5701),
				UUID:    hzTypes.NewUUID(),
			}
			hzClient.MemberMap[existingMember.String()] = true

			stateChanged := cluster.MembershipStateChanged{
				Member: existingMember,
				State:  cluster.MembershipStateRemoved,
			}
			go getStatusUpdateListener(hzClient)(stateChanged)

			Eventually(func() bool {
				return hzClient.MemberMap[stateChanged.Member.String()]
			}, timeout, interval).Should(BeFalse())

			Eventually(func() event.GenericEvent {
				select {
				case e := <-hzClient.memberEventsChannel:
					return e
				default:
					return event.GenericEvent{}
				}
			}).Should(Equal(event.GenericEvent{
				Object: &v1alpha1.Hazelcast{ObjectMeta: metav1.ObjectMeta{
					Namespace: hzClient.NamespacedName.Namespace,
					Name:      hzClient.NamespacedName.Name,
				}}}))
		})

		It("Should send event to the channel when calling triggerReconcile", func() {
			go hzClient.triggerReconcile()

			Eventually(func() event.GenericEvent {
				select {
				case e := <-hzClient.memberEventsChannel:
					return e
				default:
					return event.GenericEvent{}
				}
			}, timeout, interval).Should(Equal(event.GenericEvent{
				Object: &v1alpha1.Hazelcast{ObjectMeta: metav1.ObjectMeta{
					Namespace: hzClient.NamespacedName.Namespace,
					Name:      hzClient.NamespacedName.Name,
				}}}))
		})

		It("Should run periodic jobs for updating the member status and send events to trigger reconcile", func() {
			ctx := context.Background()
			req := testcontainers.ContainerRequest{
				Image:        "hazelcast/hazelcast",
				ExposedPorts: []string{"5701/tcp"},
				WaitingFor:   wait.ForLog("is STARTED").WithPollInterval(1 * time.Second),
			}

			hzC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err != nil {
				Fail("Unable to start Hazelcast container")
			}
			//goland:noinspection ALL
			defer hzC.Terminate(ctx)

			//goland:noinspection ALL
			ip, err := hzC.Host(ctx)
			if err != nil {
				Fail(err.Error())
			}
			port, err := hzC.MappedPort(ctx, "5701")
			if err != nil {
				Fail(err.Error())
			}

			config := hazelcast.Config{}
			//goland:noinspection ALL
			config.Cluster.Network.SetAddresses(ip + ":" + port.Port())
			err = hzClient.start(ctx, config)
			if err != nil {
				Fail("Cannot start Hazelcast client: " + err.Error())
			}

			Eventually(func() event.GenericEvent {
				select {
				case e := <-hzClient.memberEventsChannel:
					return e
				default:
					return event.GenericEvent{}
				}
			}, timeout, interval).Should(Equal(event.GenericEvent{
				Object: &v1alpha1.Hazelcast{ObjectMeta: metav1.ObjectMeta{
					Namespace: hzClient.NamespacedName.Namespace,
					Name:      hzClient.NamespacedName.Name,
				}}}))
		})
	})
})
