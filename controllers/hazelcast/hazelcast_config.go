//go:build !localrun
// +build !localrun

package hazelcast

import (
	"fmt"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
)

func buildConfig(h *hazelcastv1alpha1.Hazelcast) hazelcast.Config {
	config := hazelcast.Config{}
	cc := &config.Cluster
	cc.Name = h.Spec.ClusterName
	cc.Network.SetAddresses(fmt.Sprintf("%s.%s.svc.cluster.local:5701", h.Name, h.Namespace))
	return config
}
