package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync/atomic"
)

type HazelcastClient struct {
	Client         *hazelcast.Client
	NamespacedName types.NamespacedName
	Log            logr.Logger
}

func (c HazelcastClient) Shutdown(ctx context.Context) error {
	err := c.Client.Shutdown(ctx)
	return err
}

func NewHazelcast(ctx context.Context, config hazelcast.Config, l logr.Logger, n types.NamespacedName) (HazelcastClient, error) {
	c, err := hazelcast.StartNewClientWithConfig(ctx, config)
	if err != nil {
		return HazelcastClient{}, err
	}
	return HazelcastClient{
		Client:         c,
		NamespacedName: n,
		Log:            l,
	}, nil
}

func getStatusUpdateListener(ctx context.Context, hzClient HazelcastClient, c client.Client) func(cluster.MembershipStateChanged) {
	return func(event cluster.MembershipStateChanged) {
		println("Entering the listener " + event.Member.String())
		h := &v1alpha1.Hazelcast{}
		println("Try to get the Hz CR " + event.Member.String())
		if err := c.Get(ctx, hzClient.NamespacedName, h); err != nil {
			hzClient.Log.Error(err, "Failed to get Hazelcast")
			return
		}
		println("Go the Hz CR " + event.Member.String())
		if event.State == cluster.MembershipStateAdded {
			atomic.AddInt32(&h.Status.Cluster.CurrentMembers, 1)
		} else if event.State == cluster.MembershipStateRemoved {
			atomic.AddInt32(&h.Status.Cluster.CurrentMembers, -1)
		}
		println("Updating the CR status " + event.Member.String())
		if err := c.Status().Update(ctx, h); err != nil {
			hzClient.Log.Error(err, "Error while updating Hazelcast status")
		}
		println("Updated the CR status " + event.Member.String())
		log := fmt.Sprintf("Curent cluster members %d", h.Status.Cluster.CurrentMembers)
		println(log)
	}
}
