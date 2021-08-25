package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type HazelcastClient struct {
	Client         *hazelcast.Client
	NamespacedName types.NamespacedName
	Log            logr.Logger
	MemberMap      map[string]bool
}

func (c HazelcastClient) Shutdown(ctx context.Context) error {
	err := c.Client.Shutdown(ctx)
	return err
}

func NewHazelcastClient(l logr.Logger, n types.NamespacedName) HazelcastClient {
	return HazelcastClient{
		NamespacedName: n,
		Log:            l,
		MemberMap:      make(map[string]bool),
	}
}

func (c HazelcastClient) start(ctx context.Context, config hazelcast.Config) error {
	hzClient, err := hazelcast.StartNewClientWithConfig(ctx, config)
	if err != nil {
		return err
	}
	c.Client = hzClient
	return nil
}

func getStatusUpdateListener(hzClient HazelcastClient, memberEventChannel chan event.GenericEvent) func(cluster.MembershipStateChanged) {
	return func(changed cluster.MembershipStateChanged) {
		println("Entering the listener " + changed.Member.String())
		if changed.State == cluster.MembershipStateAdded {
			hzClient.MemberMap[changed.Member.String()] = true
			println("memberEvent: Added " + changed.Member.String())
		} else if changed.State == cluster.MembershipStateRemoved {
			delete(hzClient.MemberMap, changed.Member.String())
			println("memberEvent: Removed " + changed.Member.String())
		}
		memberEventChannel <- event.GenericEvent{
			Object: &v1alpha1.Hazelcast{ObjectMeta: metav1.ObjectMeta{
				Namespace: hzClient.NamespacedName.Namespace,
				Name:      hzClient.NamespacedName.Name,
			}}}
		println("memberEvent sent to channel " + changed.Member.String())
	}
}
