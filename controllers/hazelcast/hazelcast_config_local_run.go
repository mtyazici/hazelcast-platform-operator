//go:build localrun
// +build localrun
// This file is used  for running operator locally and connect to the cluster that uses ExposeExternally feature

package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
)

func buildConfig(_ *hazelcastv1alpha1.Hazelcast) hazelcast.Config {
	config := hazelcast.Config{}
	cc := &config.Cluster
	cc.Network.SetAddresses("127.0.0.1:8000")
	cc.Unisocket = true
	return config
}
