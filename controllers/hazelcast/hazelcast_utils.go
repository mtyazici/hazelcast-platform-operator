//go:build !kind
// +build !kind

package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func hazelcastService(h *hazelcastv1alpha1.Hazelcast) *v1.Service {
	return &corev1.Service{
		ObjectMeta: metadata(h),
		Spec: corev1.ServiceSpec{
			Selector: labels(h),
			Ports:    ports(),
		},
	}
}

func servicePerPod(i int, h *hazelcastv1alpha1.Hazelcast) *v1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      servicePerPodName(i, h),
			Namespace: h.Namespace,
			Labels:    servicePerPodLabels(h),
		},
		Spec: corev1.ServiceSpec{
			Selector:                 servicePerPodSelector(i, h),
			Ports:                    ports(),
			PublishNotReadyAddresses: true,
		},
	}
}

func ports() []v1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "hazelcast-port",
			Protocol:   corev1.ProtocolTCP,
			Port:       5701,
			TargetPort: intstr.FromString("hazelcast"),
		},
	}
}
