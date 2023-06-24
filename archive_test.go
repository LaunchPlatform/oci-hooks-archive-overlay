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
		}}, map[string]Archive{
			"/path/to/mount-point": {Name: "data", MountPoint: "/path/to/mount-point", ArchiveTo: "/path/to/archive-to"},
		},
		},
		{
			"archive-success", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point":     "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":      "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-success": "/path/to/archive-success",
		}}, map[string]Archive{
			"/path/to/mount-point": {
				Name:           "data",
				MountPoint:     "/path/to/mount-point",
				ArchiveTo:      "/path/to/archive-to",
				ArchiveSuccess: "/path/to/archive-success",
			},
		},
		},
		{
			"multiple", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data0.mount-point": "/path/to/mount-point0",
			"com.launchplatform.oci-hooks.archive-overlay.data0.archive-to":  "/path/to/archive-to0",
			"com.launchplatform.oci-hooks.archive-overlay.data1.mount-point": "/path/to/mount-point1",
			"com.launchplatform.oci-hooks.archive-overlay.data1.archive-to":  "/path/to/archive-to1",
		}}, map[string]Archive{
			"/path/to/mount-point0": {Name: "data0", MountPoint: "/path/to/mount-point0", ArchiveTo: "/path/to/archive-to0"},
			"/path/to/mount-point1": {Name: "data1", MountPoint: "/path/to/mount-point1", ArchiveTo: "/path/to/archive-to1"},
		},
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
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "",
		}}, map[string]Archive{},
		},
		{
			"empty-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
		}}, map[string]Archive{},
		},
		{
			"missing-archive-to", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
		}}, map[string]Archive{},
		},
		{
			"missing-mount-point", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to": "/path/to/archive-to",
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
