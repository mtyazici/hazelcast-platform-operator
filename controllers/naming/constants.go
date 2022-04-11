package naming

import (
	corev1 "k8s.io/api/core/v1"
)

// Labels and label values
const (
	// Finalizer name used by operator
	Finalizer = "hazelcast.com/finalizer"
	// LicenseDataKey is a key used in k8s secret that holds the Hazelcast license
	LicenseDataKey = "license-key"
	// LicenseKeySecret default license key secret
	LicenseKeySecret = "hazelcast-license-key"
	// ServicePerPodLabelName set to true when the service is a Service per pod
	ServicePerPodLabelName                       = "hazelcast.com/service-per-pod"
	ServicePerPodCountAnnotation                 = "hazelcast.com/service-per-pod-count"
	ExposeExternallyAnnotation                   = "hazelcast.com/expose-externally-member-access"
	LastSuccessfulSpecAnnotation                 = "hazelcast.com/last-successful-spec"
	CurrentHazelcastConfigForcingRestartChecksum = "hazelcast.com/current-hazelcast-config-forcing-restart-checksum"

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

	OperatorName         = "hazelcast-platform-operator"
	Hazelcast            = "hazelcast"
	HazelcastPortName    = "hazelcast-port"
	HazelcastStorageName = Hazelcast + "-storage"
	HazelcastMountPath   = "/data/hazelcast"

	// ManagementCenter MC name
	ManagementCenter = "management-center"
	// Mancenter MC short name
	Mancenter = "mancenter"
	// MancenterStorageName storage name for MC
	MancenterStorageName = Mancenter + "-storage"

	// PersistenceVolumeName is the name the Persistence Volume Claim used in Persistence configuration.
	PersistenceVolumeName = "hot-restart-persistence"
)

// Hazelcast default configurations
const (
	// DefaultHzPort Hazelcast default port
	DefaultHzPort = 5701
	// DefaultClusterSize default number of members of Hazelcast cluster
	DefaultClusterSize = 3
	// DefaultClusterName default name of Hazelcast cluster
	DefaultClusterName = "dev"
	// HazelcastRepo image repository for Hazelcast
	HazelcastRepo = "docker.io/hazelcast/hazelcast"
	// HazelcastEERepo image repository for Hazelcast EE
	HazelcastEERepo = "docker.io/hazelcast/hazelcast-enterprise"
	// HazelcastVersion version of Hazelcast image
	HazelcastVersion = "5.1"
	// HazelcastImagePullPolicy pull policy for Hazelcast Platform image
	HazelcastImagePullPolicy = corev1.PullIfNotPresent
)

// Management Center default configurations
const (
	// MCRepo image repository for Management Center
	MCRepo = "docker.io/hazelcast/management-center"
	// MCVersion version of Management Center image
	MCVersion = "5.1.1"
	// MCImagePullPolicy pull policy for Management Center image
	MCImagePullPolicy = corev1.PullIfNotPresent
)

// Map CR default values
const (
	MapBackupCount        = int32(1)
	MapAsyncBackupCount   = int32(0)
	MapTimeToLiveSeconds  = int32(0)
	MapMaxIdleSeconds     = int32(0)
	MapReadBackupData     = false
	MapPersistenceEnabled = false
)

// Operator Values
const (
	PhoneHomeEnabledEnv = "PHONE_HOME_ENABLED"
	PardotIDEnv         = "PARDOT_ID"
	OperatorVersionEnv  = "OPERATOR_VERSION"
	NamespaceEnv        = "NAMESPACE"
	PodNameEnv          = "POD_NAME"
)
