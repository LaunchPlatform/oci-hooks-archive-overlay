package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				ArchiveTo:  "/path/to/archive-to",
				TarUser:    -1,
				TarGroup:   -1,
			},
		},
		},
		{
			"archive-success", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.success":     "/path/to/archive-success",
		}}, map[string]Archive{
			"/path/to/mount-point": {
				Name:           "data",
				MountPoint:     "/path/to/mount-point",
				ArchiveTo:      "/path/to/archive-to",
				ArchiveSuccess: "/path/to/archive-success",
				TarUser:        -1,
				TarGroup:       -1,
			},
		},
		},
		{
			"method", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.method":      "tar.gz",
		}}, map[string]Archive{
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				ArchiveTo:  "/path/to/archive-to",
				Method:     "tar.gz",
				TarUser:    -1,
				TarGroup:   -1,
			},
		},
		},
		{
			"tar-content-owner", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point":       "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":        "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.method":            "tar.gz",
			"com.launchplatform.oci-hooks.archive-overlay.data.tar-content-owner": "2000:3000",
		}}, map[string]Archive{
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				ArchiveTo:  "/path/to/archive-to",
				Method:     "tar.gz",
				TarUser:    2000,
				TarGroup:   3000,
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
			"/path/to/mount-point0": {
				Name:       "data0",
				MountPoint: "/path/to/mount-point0",
				ArchiveTo:  "/path/to/archive-to0",
				TarUser:    -1,
				TarGroup:   -1,
			},
			"/path/to/mount-point1": {
				Name:       "data1",
				MountPoint: "/path/to/mount-point1",
				ArchiveTo:  "/path/to/archive-to1",
				TarUser:    -1,
				TarGroup:   -1,
			},
		},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.mount-point": "/path/to/mount-point",
			"com.launchplatform.oci-hooks.archive-overlay.data.archive-to":  "/path/to/archive-to",
			"com.launchplatform.oci-hooks.archive-overlay.data.invalid":     "others",
		}}, map[string]Archive{
			"/path/to/mount-point": {
				Name:       "data",
				MountPoint: "/path/to/mount-point",
				ArchiveTo:  "/path/to/archive-to",
				TarUser:    -1,
				TarGroup:   -1,
			}},
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

func Test_parseOwner(t *testing.T) {
	type args struct {
		owner string
	}
	tests := []struct {
		name    string
		args    args
		uid     int
		gid     int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"only-user", args{"2000"}, 2000, 0, assert.NoError,
		},
		{
			"both", args{"2000:3000"}, 2000, 3000, assert.NoError,
		},
		{
			"empty", args{""}, 0, 0, assert.Error,
		},
		{
			"more-than-two-parts", args{"1:2:3"}, 0, 0, assert.Error,
		},
		{
			"non-int-user", args{"user"}, 0, 0, assert.Error,
		},
		{
			"non-int-both", args{"user:group"}, 0, 0, assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseOwner(tt.args.owner)
			if !tt.wantErr(t, err, fmt.Sprintf("parseOwner(%v)", tt.args.owner)) {
				return
			}
			assert.Equalf(t, tt.uid, got, "parseOwner(%v)", tt.args.owner)
			assert.Equalf(t, tt.gid, got1, "parseOwner(%v)", tt.args.owner)
		})
	}
}
