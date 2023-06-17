package main

import (
	"reflect"
	"testing"
)

func TestParseArchives(t *testing.T) {
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
			"com.launchplatform.oci-hooks.archive-overlay.data.src":  "/path/to/src",
			"com.launchplatform.oci-hooks.archive-overlay.data.dest": "/path/to/dest",
		}}, map[string]Archive{"/path/to/dest": {Name: "data", Src: "/path/to/src", Dest: "/path/to/dest"}},
		},
		{
			"invalid-key", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.src":     "/path/to/src",
			"com.launchplatform.oci-hooks.archive-overlay.data.dest":    "/path/to/dest",
			"com.launchplatform.oci-hooks.archive-overlay.data.invalid": "others",
		}}, map[string]Archive{"/path/to/dest": {Name: "data", Src: "/path/to/src", Dest: "/path/to/dest"}},
		},
		{
			"empty-src", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.dest": "/path/to/dest",
		}}, map[string]Archive{},
		},
		{
			"empty-dest", args{annotations: map[string]string{
			"com.launchplatform.oci-hooks.archive-overlay.data.src": "/path/to/src",
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
