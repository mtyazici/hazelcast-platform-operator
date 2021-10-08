package naming

// Labels and label values
const (
	// Finalizer name used by operator
	Finalizer = "hazelcast.com/finalizer"
	// LicenseDataKey is a key used in k8s secret that holds the Hazelcast license
	LicenseDataKey = "license-key"
	// ServicePerPodLabelName
	ServicePerPodLabelName       = "hazelcast.com/service-per-pod"
	ServicePerPodCountAnnotation = "hazelcast.com/service-per-pod-count"
	ExposeExternallyAnnotation   = "hazelcast.com/expose-externally-member-access"

	// PodNameLabel label that represents the name of the pod in the StatefulSet
	PodNameLabel = "statefulset.kubernetes.io/pod-name"
	// ApplicationNameLabel label for the name of the application
	ApplicationNameLabel = "app.kubernetes.io/name"
	// ApplicationInstanceNameLabel label for a unique name identifying the instance of an application
	ApplicationInstanceNameLabel = "app.kubernetes.io/instance"
	// ApplicationManagedByLabel label for the tool being used to manage the operation of an application
	ApplicationManagedByLabel = "app.kubernetes.io/managed-by"

	LabelValueTrue  = "true"
	LabelValueFalse = "false"

	OperatorName     = "hazelcast-enterprise-operator"
	ManagementCenter = "management-center"
	Hazelcast        = "hazelcast"

	HazelcastPortName = "hazelcast-port"
)

// Environment variables used for Hazelcast cluster configuration
const (
	// KubernetesEnabled enable Kubernetes discovery
	KubernetesEnabled = "HZ_NETWORK_JOIN_KUBERNETES_ENABLED"
	// KubernetesServiceName used to scan only PODs connected to the given service
	KubernetesServiceName = "HZ_NETWORK_JOIN_KUBERNETES_SERVICENAME"
	// KubernetesNodeNameAsExternalAddress uses the node name to connect to a NodePort service instead of looking up the external IP using the API
	KubernetesNodeNameAsExternalAddress = "HZ_NETWORK_JOIN_KUBERNETES_USENODENAMEASEXTERNALADDRESS"
	// KubernetesServicePerPodLabel label name used to tag services that should form the Hazelcast cluster together
	KubernetesServicePerPodLabel = "HZ_NETWORK_JOIN_KUBERNETES_SERVICEPERPODLABELNAME"
	// KubernetesServicePerPodLabelValue label value used to tag services that should form the Hazelcast cluster together
	KubernetesServicePerPodLabelValue = "HZ_NETWORK_JOIN_KUBERNETES_SERVICEPERPODLABELVALUE"

	RESTEnabled            = "HZ_NETWORK_RESTAPI_ENABLED"
	RESTHealthCheckEnabled = "HZ_NETWORK_RESTAPI_ENDPOINTGROUPS_HEALTHCHECK_ENABLED"

	LicenseKey  = "HZ_LICENSEKEY"
	ClusterName = "HZ_CLUSTERNAME"
)

// Hazelcast default configurations
const (
	// DefaultHzPort Hazelcast default port
	DefaultHzPort = 5701
)
