package v1alpha1

import (
	"fmt"
	"hash/fnv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Phase represents the current state of the cluster
type Phase string

const (
	// Running phase is the state when all the members of the cluster are successfully started
	Running Phase = "Running"
	// Failed phase is the state of error during the cluster startup
	Failed Phase = "Failed"
	// Pending phase is the state of starting the cluster when not all the members are started yet
	Pending Phase = "Pending"
)

// HazelcastSpec defines the desired state of Hazelcast
type HazelcastSpec struct {
	// Number of Hazelcast members in the cluster.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=3
	// +optional
	ClusterSize int32 `json:"clusterSize"`

	// Repository to pull the Hazelcast Platform image from.
	// +kubebuilder:default:="docker.io/hazelcast/hazelcast-enterprise"
	// +optional
	Repository string `json:"repository"`

	// Version of Hazelcast Platform.
	// +kubebuilder:default:="5.0"
	// +optional
	Version string `json:"version"`

	// Pull policy for the Hazelcast Platform image
	// +kubebuilder:default:="IfNotPresent"
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// Name of the secret with Hazelcast Enterprise License Key.
	// +optional
	LicenseKeySecret string `json:"licenseKeySecret"`

	// Configuration to expose Hazelcast cluster to external clients.
	// +optional
	ExposeExternally ExposeExternallyConfiguration `json:"exposeExternally"`

	// Name of the Hazelcast cluster.
	// +kubebuilder:default:="dev"
	// +optional
	ClusterName string `json:"clusterName"`

	// Scheduling details
	// +optional
	Scheduling SchedulingConfiguration `json:"scheduling,omitempty"`
}

// SchedulingConfiguration defines the pods scheduling details
type SchedulingConfiguration struct {
	// Affinity
	// +optional
	Affinity corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NodeSelector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// ExposeExternallyConfiguration defines how to expose Hazelcast cluster to external clients
type ExposeExternallyConfiguration struct {
	// Specifies how members are exposed.
	// Valid values are:
	// - "Smart" (default): each member pod is exposed with a separate external address
	// - "Unisocket": all member pods are exposed with one external address
	// +optional
	Type ExposeExternallyType `json:"type,omitempty"`

	// Type of the service used to discover Hazelcast cluster.
	// +optional
	DiscoveryServiceType corev1.ServiceType `json:"discoveryServiceType,omitempty"`

	// How each member is accessed from the external client.
	// Valid values are:
	// - "NodePortExternalIP" (default): each member is accessed by the NodePort service and the node external IP/hostname
	// - "NodePortNodeName": each member is accessed by the NodePort service and the node name
	// - "LoadBalancer": each member is accessed by the LoadBalancer service external address
	// +optional
	MemberAccess MemberAccess `json:"memberAccess,omitempty"`
}

// ExposeExternallyType describes how Hazelcast members are exposed.
// +kubebuilder:validation:Enum=Smart;Unisocket
type ExposeExternallyType string

const (
	// ExposeExternallyTypeSmart exposes each Hazelcast member with a separate external address.
	ExposeExternallyTypeSmart ExposeExternallyType = "Smart"

	// ExposeExternallyTypeUnisocket exposes all Hazelcast members with one external address.
	ExposeExternallyTypeUnisocket ExposeExternallyType = "Unisocket"
)

// MemberAccess describes how each Hazelcast member is accessed from the external client.
// +kubebuilder:validation:Enum=NodePortExternalIP;NodePortNodeName;LoadBalancer
type MemberAccess string

const (
	// MemberAccessNodePortExternalIP lets the client access Hazelcast member with the NodePort service and the node external IP/hostname
	MemberAccessNodePortExternalIP MemberAccess = "NodePortExternalIP"

	// MemberAccessNodePortNodeName lets the client access Hazelcast member with the NodePort service and the node name
	MemberAccessNodePortNodeName MemberAccess = "NodePortNodeName"

	// MemberAccessLoadBalancer lets the client access Hazelcast member with the LoadBalancer service
	MemberAccessLoadBalancer MemberAccess = "LoadBalancer"
)

// Returns true if exposeExternally configuration is specified.
func (c *ExposeExternallyConfiguration) IsEnabled() bool {
	return !(*c == (ExposeExternallyConfiguration{}))
}

// Returns true if Smart configuration is specified and therefore each Hazelcast member needs to be exposed with a separate address.
func (c *ExposeExternallyConfiguration) IsSmart() bool {
	return c.Type == ExposeExternallyTypeSmart
}

// Returns true if Hazelcast client wants to use Node Name instead of External IP.
func (c *ExposeExternallyConfiguration) UsesNodeName() bool {
	return c.MemberAccess == MemberAccessNodePortNodeName
}

// Returns service type that is used for the cluster discovery (LoadBalancer by default).
func (c *ExposeExternallyConfiguration) DiscoveryK8ServiceType() corev1.ServiceType {
	switch c.DiscoveryServiceType {
	case corev1.ServiceTypeNodePort:
		return corev1.ServiceTypeNodePort
	default:
		return corev1.ServiceTypeLoadBalancer
	}
}

// Returns service type that is used for the communication with each member (NodePort by default).
func (c *ExposeExternallyConfiguration) MemberAccessServiceType() corev1.ServiceType {
	switch c.MemberAccess {
	case MemberAccessLoadBalancer:
		return corev1.ServiceTypeLoadBalancer
	default:
		return corev1.ServiceTypeNodePort
	}
}

// HazelcastStatus defines the observed state of Hazelcast
type HazelcastStatus struct {
	Phase   Phase                  `json:"phase"`
	Cluster HazelcastClusterStatus `json:"hazelcastClusterStatus"`
	Message string                 `json:"message,omitempty"`
}

// HazelcastClusterStatus defines the status of the Hazelcast cluster
type HazelcastClusterStatus struct {
	// ReadyMembers represents the number of members that are connected to cluster from the desired number of members
	// in the format <ready>/<desired>
	ReadyMembers string `json:"readyMembers"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Hazelcast is the Schema for the hazelcasts API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Current state of the Hazelcast deployment"
// +kubebuilder:printcolumn:name="Members",type="string",JSONPath=".status.hazelcastClusterStatus.readyMembers",description="Current numbers of ready Hazelcast members"
type Hazelcast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HazelcastSpec   `json:"spec,omitempty"`
	Status HazelcastStatus `json:"status,omitempty"`
}

func (h *Hazelcast) DockerImage() string {
	return fmt.Sprintf("%s:%s", h.Spec.Repository, h.Spec.Version)
}

//+kubebuilder:object:root=true

// HazelcastList contains a list of Hazelcast
type HazelcastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hazelcast `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hazelcast{}, &HazelcastList{})
}

func FNV32a(txt string) uint32 {
	alg := fnv.New32a()
	alg.Write([]byte(txt))
	return alg.Sum32()
}

func (h *Hazelcast) ClusterScopedName() string {
	return fmt.Sprintf("%s-%d", h.Name, FNV32a(h.Namespace))
}
