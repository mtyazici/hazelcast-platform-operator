//go:build !localrun
// +build !localrun

package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
)

func buildConfig(h *hazelcastv1alpha1.Hazelcast) hazelcast.Config {
	config := hazelcast.Config{}
	config.Cluster.Network.SetAddresses(h.Name + ":5701")
	return config
}
