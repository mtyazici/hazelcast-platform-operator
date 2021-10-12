package hazelcast

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hazelcast/hazelcast-enterprise-operator/controllers/naming"

	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-enterprise-operator/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *HazelcastReconciler) addFinalizer(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	if !controllerutil.ContainsFinalizer(h, naming.Finalizer) {
		controllerutil.AddFinalizer(h, naming.Finalizer)
		err := r.Update(ctx, h)
		if err != nil {
			logger.Error(err, "Failed to add finalizer into custom resource")
			return err
		}
		logger.V(1).Info("Finalizer added into custom resource successfully")
	}
	return nil
}

func (r *HazelcastReconciler) executeFinalizer(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	if err := r.removeClusterRole(ctx, h, logger); err != nil {
		logger.Error(err, "ClusterRole could not be removed")
		return err
	}
	if err := r.removeClusterRoleBinding(ctx, h, logger); err != nil {
		logger.Error(err, "ClusterRoleBinding could not be removed")
		return err
	}
	controllerutil.RemoveFinalizer(h, naming.Finalizer)
	err := r.Update(ctx, h)
	if err != nil {
		logger.Error(err, "Failed to remove finalizer from custom resource")
		return err
	}
	key := types.NamespacedName{Name: h.Name, Namespace: h.Namespace}
	if c, ok := r.hzClients.Load(key); ok {
		r.hzClients.Delete(key)
		// shutdown error is ignored and does not need to be handled
		_ = c.(*HazelcastClient).Client.Shutdown(ctx)
	}
	return nil
}

func (r *HazelcastReconciler) removeClusterRole(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	clusterRole := &rbacv1.ClusterRole{}
	err := r.Get(ctx, client.ObjectKey{Name: h.Name}, clusterRole)
	if err != nil && errors.IsNotFound(err) {
		logger.V(1).Info("ClusterRole is not created yet. Or it is already removed.")
		return nil
	}

	err = r.Delete(ctx, clusterRole)
	if err != nil {
		logger.Error(err, "Failed to clean up ClusterRole")
		return err
	}
	logger.V(1).Info("ClusterRole removed successfully")
	return nil
}

func (r *HazelcastReconciler) removeClusterRoleBinding(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	crb := &rbacv1.ClusterRoleBinding{}
	err := r.Get(ctx, client.ObjectKey{Name: h.Name}, crb)
	if err != nil && errors.IsNotFound(err) {
		logger.V(1).Info("ClusterRoleBinding is not created yet. Or it is already removed.")
		return nil
	}

	err = r.Delete(ctx, crb)
	if err != nil {
		logger.Error(err, "Failed to clean up ClusterRoleBinding")
		return err
	}
	logger.V(1).Info("ClusterRoleBinding removed successfully")
	return nil
}

func (r *HazelcastReconciler) reconcileClusterRole(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   h.Name,
			Labels: labels(h),
		},
	}

	opResult, err := util.CreateOrUpdate(ctx, r.Client, clusterRole, func() error {
		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints", "pods", "nodes", "services"},
				Verbs:     []string{"get", "list"},
			},
		}
		return nil
	})
	if opResult != controllerutil.OperationResultNone {
		logger.Info("Operation result", "ClusterRole", h.Name, "result", opResult)
	}
	return err
}

func (r *HazelcastReconciler) reconcileServiceAccount(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metadata(h),
	}

	err := controllerutil.SetControllerReference(h, serviceAccount, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set owner reference on ServiceAccount")
		return err
	}

	opResult, err := util.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
		return nil
	})
	if opResult != controllerutil.OperationResultNone {
		logger.Info("Operation result", "ServiceAccount", h.Name, "result", opResult)
	}
	return err
}

func (r *HazelcastReconciler) reconcileClusterRoleBinding(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   h.Name,
			Labels: labels(h),
		},
	}

	opResult, err := util.CreateOrUpdate(ctx, r.Client, crb, func() error {
		crb.Subjects = []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      h.Name,
				Namespace: h.Namespace,
			},
		}
		crb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     h.Name,
		}

		return nil
	})
	if opResult != controllerutil.OperationResultNone {
		logger.Info("Operation result", "RoleBinding", h.Name, "result", opResult)
	}
	return err
}

func (r *HazelcastReconciler) reconcileService(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	service := &corev1.Service{
		ObjectMeta: metadata(h),
		Spec: corev1.ServiceSpec{
			Selector: labels(h),
			Ports:    ports(),
		},
	}

	err := controllerutil.SetControllerReference(h, service, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set owner reference on Service")
		return err
	}

	opResult, err := util.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec.Type = serviceType(h)
		if serviceType(h) == corev1.ServiceTypeClusterIP {
			// dirty hack to prevent the error when changing the service type
			service.Spec.Ports[0].NodePort = 0
		}

		return nil
	})
	if opResult != controllerutil.OperationResultNone {
		logger.Info("Operation result", "Service", h.Name, "result", opResult)
	}
	return err
}

func serviceType(h *hazelcastv1alpha1.Hazelcast) v1.ServiceType {
	if h.Spec.ExposeExternally.IsEnabled() {
		return h.Spec.ExposeExternally.DiscoveryK8ServiceType()
	}
	return corev1.ServiceTypeClusterIP
}

func (r *HazelcastReconciler) reconcileServicePerPod(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	if !h.Spec.ExposeExternally.IsSmart() {
		// Service per pod applies only to Smart type
		return nil
	}

	for i := 0; i < int(h.Spec.ClusterSize); i++ {
		service := &corev1.Service{
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

		err := controllerutil.SetControllerReference(h, service, r.Scheme)
		if err != nil {
			return err
		}

		opResult, err := util.CreateOrUpdate(ctx, r.Client, service, func() error {
			service.Spec.Type = h.Spec.ExposeExternally.MemberAccessServiceType()
			return nil
		})

		if opResult != controllerutil.OperationResultNone {
			logger.Info("Operation result", "Service", servicePerPodName(i, h), "result", opResult)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *HazelcastReconciler) reconcileUnusedServicePerPod(ctx context.Context, h *hazelcastv1alpha1.Hazelcast) error {
	var n int
	if h.Spec.ExposeExternally.IsSmart() {
		n = int(h.Spec.ClusterSize)
	}

	// Delete unused services (when the cluster was scaled down)
	// The current number of service per pod is always stored in the StatefulSet annotations
	sts := &appsv1.StatefulSet{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: h.Name, Namespace: h.Namespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Not found, StatefulSet is not created yet, no need to delete any services
			return nil
		}
		return err
	}
	p, err := strconv.Atoi(sts.ObjectMeta.Annotations[naming.ServicePerPodCountAnnotation])
	if err != nil {
		// Annotation not found, no need to delete any services
		return nil
	}

	for i := n; i < p; i++ {
		s := &v1.Service{}
		err := r.Client.Get(ctx, client.ObjectKey{Name: servicePerPodName(i, h), Namespace: h.Namespace}, s)
		if err != nil {
			if errors.IsNotFound(err) {
				// Not found, no need to remove the service
				continue
			}
			return err
		}
		err = r.Client.Delete(ctx, s)
		if err != nil {
			if errors.IsNotFound(err) {
				// Not found, no need to remove the service
				continue
			}
			return err
		}
	}

	return nil
}

func servicePerPodName(i int, h *hazelcastv1alpha1.Hazelcast) string {
	return fmt.Sprintf("%s-%d", h.Name, i)
}

func servicePerPodSelector(i int, h *hazelcastv1alpha1.Hazelcast) map[string]string {
	ls := labels(h)
	ls[naming.PodNameLabel] = servicePerPodName(i, h)
	return ls
}

func servicePerPodLabels(h *hazelcastv1alpha1.Hazelcast) map[string]string {
	ls := labels(h)
	ls[naming.ServicePerPodLabelName] = naming.LabelValueTrue
	return ls
}

func ports() []v1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       naming.HazelcastPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       naming.DefaultHzPort,
			TargetPort: intstr.FromString(naming.Hazelcast),
		},
	}
}

func (r *HazelcastReconciler) isServicePerPodReady(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, _ logr.Logger) bool {
	if !h.Spec.ExposeExternally.IsSmart() {
		// Service per pod applies only to Smart type
		return true
	}

	// Check if each service per pod is ready
	for i := 0; i < int(h.Spec.ClusterSize); i++ {
		s := &v1.Service{}
		err := r.Client.Get(ctx, client.ObjectKey{Name: servicePerPodName(i, h), Namespace: h.Namespace}, s)
		if err != nil {
			// Service is not created yet
			return false
		}
		if s.Spec.Type == v1.ServiceTypeLoadBalancer {
			if len(s.Status.LoadBalancer.Ingress) == 0 {
				// LoadBalancer service waiting for External IP to get assigned
				return false
			}
		}
	}

	return true
}

func (r *HazelcastReconciler) reconcileStatefulset(ctx context.Context, h *hazelcastv1alpha1.Hazelcast, logger logr.Logger) error {
	ls := labels(h)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metadata(h),
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			ServiceName: h.Name,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: h.Name,
					Containers: []v1.Container{{
						Name: naming.Hazelcast,
						Ports: []v1.ContainerPort{{
							ContainerPort: naming.DefaultHzPort,
							Name:          naming.Hazelcast,
							Protocol:      v1.ProtocolTCP,
						}},
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path:   "/hazelcast/health/node-state",
									Port:   intstr.FromInt(naming.DefaultHzPort),
									Scheme: corev1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: 0,
							TimeoutSeconds:      10,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    10,
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{
									Path:   "/hazelcast/health/node-state",
									Port:   intstr.FromInt(naming.DefaultHzPort),
									Scheme: corev1.URISchemeHTTP,
								},
							},
							InitialDelaySeconds: 0,
							TimeoutSeconds:      10,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							FailureThreshold:    10,
						},
						SecurityContext: &v1.SecurityContext{
							RunAsNonRoot:             &[]bool{true}[0],
							RunAsUser:                &[]int64{65534}[0],
							Privileged:               &[]bool{false}[0],
							ReadOnlyRootFilesystem:   &[]bool{true}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &v1.Capabilities{
								Drop: []v1.Capability{"ALL"},
							},
						},
					}},
					TerminationGracePeriodSeconds: &[]int64{600}[0],
				},
			},
		},
	}

	err := controllerutil.SetControllerReference(h, sts, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set owner reference on Statefulset")
		return err
	}

	opResult, err := util.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec.Replicas = &h.Spec.ClusterSize
		sts.ObjectMeta.Annotations = statefulSetAnnotations(h)
		sts.Spec.Template.Annotations = podAnnotations(h)
		sts.Spec.Template.Spec.Containers[0].Image = h.DockerImage()
		sts.Spec.Template.Spec.Containers[0].Env = env(h)
		return nil
	})
	if opResult != controllerutil.OperationResultNone {
		logger.Info("Operation result", "Statefulset", h.Name, "result", opResult)
	}
	return err
}

func env(h *hazelcastv1alpha1.Hazelcast) []v1.EnvVar {
	envs := []v1.EnvVar{
		{
			Name: naming.HzLicenseKey,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: h.Spec.LicenseKeySecret,
					},
					Key: naming.LicenseDataKey,
				},
			},
		},
		{Name: naming.KubernetesEnabled, Value: naming.LabelValueTrue},
		{Name: naming.KubernetesServiceName, Value: h.Name},
		{Name: naming.RESTEnabled, Value: naming.LabelValueTrue},
		{Name: naming.RESTHealthCheckEnabled, Value: naming.LabelValueTrue},
	}

	if h.Spec.ExposeExternally.UsesNodeName() {
		envs = append(envs, v1.EnvVar{Name: naming.KubernetesNodeNameAsExternalAddress, Value: naming.LabelValueTrue})
	}

	if h.Spec.ExposeExternally.IsSmart() {
		envs = append(envs,
			v1.EnvVar{Name: naming.KubernetesServicePerPodLabel, Value: naming.ServicePerPodLabelName},
			v1.EnvVar{Name: naming.KubernetesServicePerPodLabelValue, Value: naming.LabelValueTrue},
		)
	}

	if h.Spec.ClusterName != "" {
		envs = append(envs, v1.EnvVar{Name: naming.ClusterName, Value: h.Spec.ClusterName})
	}

	return envs
}

func labels(h *hazelcastv1alpha1.Hazelcast) map[string]string {
	return map[string]string{
		naming.ApplicationNameLabel:         naming.Hazelcast,
		naming.ApplicationInstanceNameLabel: h.Name,
		naming.ApplicationManagedByLabel:    naming.OperatorName,
	}
}

func statefulSetAnnotations(h *hazelcastv1alpha1.Hazelcast) map[string]string {
	ans := map[string]string{}
	if h.Spec.ExposeExternally.IsSmart() {
		ans[naming.ServicePerPodCountAnnotation] = strconv.Itoa(int(h.Spec.ClusterSize))
	}
	return ans
}

func podAnnotations(h *hazelcastv1alpha1.Hazelcast) map[string]string {
	ans := map[string]string{}
	if h.Spec.ExposeExternally.IsSmart() {
		ans[naming.ExposeExternallyAnnotation] = string(h.Spec.ExposeExternally.MemberAccess)
	}
	return ans
}

func metadata(h *hazelcastv1alpha1.Hazelcast) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      h.Name,
		Namespace: h.Namespace,
		Labels:    labels(h),
	}
}
