package hazelcast

import (
	"testing"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/controllers/naming"
)

func Test_mergeHazelcastSpecs(t *testing.T) {
	type args struct {
		source *hazelcastv1alpha1.HazelcastSpec
	}
	tests := []struct {
		name   string
		target *hazelcastv1alpha1.HazelcastSpec
		want   hazelcastv1alpha1.HazelcastSpec
	}{
		{
			name:   "Empty hazelcast repository",
			target: &hazelcastv1alpha1.HazelcastSpec{},
			want:   hazelcastv1alpha1.HazelcastSpec{Repository: n.HazelcastRepo},
		},
		{
			name:   "Empty hazelcast version",
			target: &hazelcastv1alpha1.HazelcastSpec{},
			want:   hazelcastv1alpha1.HazelcastSpec{Version: n.HazelcastVersion},
		},
		{
			name:   "Empty license key secret",
			target: &hazelcastv1alpha1.HazelcastSpec{},
			want:   hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: n.LicenseKeySecret},
		},
		{
			name:   "Empty cluster size",
			target: &hazelcastv1alpha1.HazelcastSpec{},
			want:   hazelcastv1alpha1.HazelcastSpec{ClusterSize: n.DefaultClusterSize},
		},
		{
			name:   "Non empty hazelcast repository",
			target: &hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast"},
			want:   hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast"},
		},
		{
			name:   "Non empty hazelcast version",
			target: &hazelcastv1alpha1.HazelcastSpec{Version: "4.2"},
			want:   hazelcastv1alpha1.HazelcastSpec{Version: "4.2"},
		},
		{
			name:   "Non empty license key secret",
			target: &hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret"},
			want:   hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret"},
		},
		{
			name:   "Non empty cluster size",
			target: &hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5},
			want:   hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyDefaultHazelcastSpecs(tt.target)
			if *tt.target != tt.want {
				t.Errorf("HazelcastSpec = %v, want %v", tt.target, tt.want)
			}
		})
	}
}
