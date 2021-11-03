package hazelcast

import (
	"testing"

	"github.com/hazelcast/hazelcast-enterprise-operator/controllers/naming"

	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
)

func Test_mergeHazelcastSpecs(t *testing.T) {
	type args struct {
		target *hazelcastv1alpha1.HazelcastSpec
		source *hazelcastv1alpha1.HazelcastSpec
	}
	tests := []struct {
		name string
		args args
		want hazelcastv1alpha1.HazelcastSpec
	}{
		{
			name: "Empty hazelcast repository",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{},
				source: &hazelcastv1alpha1.HazelcastSpec{Repository: naming.HazelcastRepo},
			},
			want: hazelcastv1alpha1.HazelcastSpec{Repository: naming.HazelcastRepo},
		},
		{
			name: "Empty hazelcast version",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{},
				source: &hazelcastv1alpha1.HazelcastSpec{Version: naming.HazelcastVersion},
			},
			want: hazelcastv1alpha1.HazelcastSpec{Version: naming.HazelcastVersion},
		},
		{
			name: "Empty license key secret",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{},
				source: &hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: naming.LicenseKeySecret},
			},
			want: hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: naming.LicenseKeySecret},
		},
		{
			name: "Empty cluster size",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{},
				source: &hazelcastv1alpha1.HazelcastSpec{ClusterSize: naming.DefaultClusterSize},
			},
			want: hazelcastv1alpha1.HazelcastSpec{ClusterSize: naming.DefaultClusterSize},
		},
		{
			name: "Non empty hazelcast repository",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast"},
				source: &hazelcastv1alpha1.HazelcastSpec{Repository: naming.HazelcastRepo},
			},
			want: hazelcastv1alpha1.HazelcastSpec{Repository: "myorg/hazelcast"},
		},
		{
			name: "Non empty hazelcast version",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{Version: "4.2"},
				source: &hazelcastv1alpha1.HazelcastSpec{Version: naming.HazelcastVersion},
			},
			want: hazelcastv1alpha1.HazelcastSpec{Version: "4.2"},
		},
		{
			name: "Non empty license key secret",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret"},
				source: &hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: naming.LicenseKeySecret},
			},
			want: hazelcastv1alpha1.HazelcastSpec{LicenseKeySecret: "license-key-secret"},
		},
		{
			name: "Non empty cluster size",
			args: args{
				target: &hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5},
				source: &hazelcastv1alpha1.HazelcastSpec{ClusterSize: naming.DefaultClusterSize},
			},
			want: hazelcastv1alpha1.HazelcastSpec{ClusterSize: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeHazelcastSpecs(tt.args.target, tt.args.source)
			if *tt.args.target != tt.want {
				t.Errorf("HazelcastSpec = %v, want %v", tt.args.target, tt.want)
			}
		})
	}
}
