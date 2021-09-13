//go:build localrun
// +build localrun
// This file is used  for running operator locally and connect to the cluster that uses ExposeExternally feature

package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-go-client"
	"os"
)

func buildConfig(h *hazelcastv1alpha1.Hazelcast) hazelcast.Config {
	if !h.Spec.ExposeExternally.IsEnabled() {
		panic("It's required to enable ExposeExternally feature to run of the operator locally.")
	}
	url := os.Getenv("DISCOVERY_SERVICE_URL")
	if url == "" {
		panic("It's required to set DISCOVERY_SERVICE_URL env to run of the operator locally.")
	}
	config := hazelcast.Config{}
	cc := &config.Cluster
	cc.Network.SetAddresses(url)
	if h.Spec.ExposeExternally.Type == hazelcastv1alpha1.ExposeExternallyTypeSmart {
		cc.Discovery.UsePublicIP = true
	}
	if h.Spec.ExposeExternally.Type == hazelcastv1alpha1.ExposeExternallyTypeUnisocket {
		cc.Unisocket = true
	}
	return config
}
