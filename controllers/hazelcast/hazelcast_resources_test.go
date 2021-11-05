package hazelcast

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/rbac/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/controllers/naming"
)

func Test_mergeHazelcastSpecs(t *testing.T) {
	defaultHzSpec := hazelcastv1alpha1.HazelcastSpec{
		ClusterSize:      n.DefaultClusterSize,
		Version:          n.HazelcastVersion,
		LicenseKeySecret: n.LicenseKeySecret,
		Repository:       n.HazelcastRepo,
	}
	tests := []struct {
		name   string
		target hazelcastv1alpha1.HazelcastSpec
		want   hazelcastv1alpha1.HazelcastSpec
	}{
		{
			name:   "Empty hazelcast repository",
			target: hazelcastv1alpha1.HazelcastSpec{ClusterSize: n.DefaultClusterSize, Version: n.HazelcastVersion, LicenseKeySecret: n.LicenseKeySecret},
			want:   defaultHzSpec,
		},
		{
			name:   "Empty hazelcast version",
			target: hazelcastv1alpha1.HazelcastSpec{ClusterSize: n.DefaultClusterSize, Repository: n.HazelcastRepo, LicenseKeySecret: n.LicenseKeySecret},
			want:   defaultHzSpec,
		},
		{
			name:   "Empty license key secret",
			target: hazelcastv1alpha1.HazelcastSpec{ClusterSize: n.DefaultClusterSize, Repository: n.HazelcastRepo, Version: n.HazelcastVersion},
			want:   defaultHzSpec,
		},
		{
			name:   "Empty cluster size",
			target: hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: n.LicenseKeySecret, Repository: n.HazelcastRepo, Version: n.HazelcastVersion},
			want:   defaultHzSpec,
		},
		{
			name:   "Non empty hazelcast repository",
			target: hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast"},
			want:   hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast", ClusterSize: n.DefaultClusterSize, Version: n.HazelcastVersion, LicenseKeySecret: n.LicenseKeySecret},
		},
		{
			name:   "Non empty hazelcast version",
			target: hazelcastv1alpha1.HazelcastSpec{Version: "4.2"},
			want:   hazelcastv1alpha1.HazelcastSpec{Version: "4.2", ClusterSize: n.DefaultClusterSize, Repository: n.HazelcastRepo, LicenseKeySecret: n.LicenseKeySecret},
		},
		{
			name:   "Non empty license key secret",
			target: hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret"},
			want:   hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret", ClusterSize: n.DefaultClusterSize, Repository: n.HazelcastRepo, Version: n.HazelcastVersion},
		},
		{
			name:   "Non empty cluster size",
			target: hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5},
			want:   hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5, LicenseKeySecret: n.LicenseKeySecret, Repository: n.HazelcastRepo, Version: n.HazelcastVersion},
		},
	}
	scheme, _ := hazelcastv1alpha1.SchemeBuilder.Register(&hazelcastv1alpha1.Hazelcast{}, &hazelcastv1alpha1.HazelcastList{}).Build()
	h := &hazelcastv1alpha1.Hazelcast{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hazelcast",
			Namespace: "default",
		},
	}
	r := HazelcastReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(h).Build(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.Spec = tt.target
			err := r.applyDefaultHazelcastSpecs(context.Background(), h)
			if err != nil {
				t.Errorf("Unexpected error have occured: %v", err)
			}
			if h.Spec != tt.want {
				t.Errorf("HazelcastSpec = %v, want %v", tt.target, tt.want)
			}
		})
	}
}

func Test_clientShutdownWhenConnectionNotEstablished(t *testing.T) {
	scheme, _ := hazelcastv1alpha1.SchemeBuilder.
		Register(&hazelcastv1alpha1.Hazelcast{}, &hazelcastv1alpha1.HazelcastList{}, &v1.ClusterRole{}, &v1.ClusterRoleBinding{}).
		Build()
	h := &hazelcastv1alpha1.Hazelcast{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hazelcast",
			Namespace: "default",
		},
	}
	r := HazelcastReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(h).Build(),
	}
	r.hzClients.Store(types.NamespacedName{Name: h.Name, Namespace: h.Namespace}, &HazelcastClient{})

	err := r.executeFinalizer(context.Background(), h, ctrl.Log)
	if err != nil {
		t.Errorf("Error while executing finilazer: %v.", err)
	}
}
