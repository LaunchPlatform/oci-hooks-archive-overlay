package main

import (
	"reflect"
	"testing"
)

func Test_parseArchives(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]Archive
	}{
		{"empty", args{annotations: map[string]string{"foo": "bar"}}, map[string]Archive{}},
		{
			"one", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
		}}, map[string]Archive{"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", ArchiveTo: "/path/to/archive-to"}},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.invalid":     "others",
		}}, map[string]Archive{"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", ArchiveTo: "/path/to/archive-to"}},
		},
		{
			"empty-archive-to", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to": "/path/to/archive-to",
		}}, map[string]Archive{},
		},
		{
			"empty-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
		}}, map[string]Archive{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseArchives(tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArchives() = %v, want %v", got, tt.want)
			}
		})
	}
}
