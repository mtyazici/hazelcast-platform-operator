package config

type HazelcastClientWrapper struct {
	HazelcastClient HazelcastClient `yaml:"hazelcast-client"`
}
type HazelcastClient struct {
	Network     NetworkClient `yaml:"network,omitempty"`
	ClusterName string        `yaml:"cluster-name,omitempty"`
}

type NetworkClient struct {
	Kubernetes Kubernetes `yaml:"kubernetes,omitempty"`
}
