package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_loadSpec(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bundle")
	if err != nil {
		t.Fatal(err)
	}
	specValue := spec.Spec{
		Version: spec.Version,
		Mounts: []spec.Mount{
			{
				Destination: "/data",
				Source:      "/path/to/source",
				Options:     []string{"nodev"},
			},
		},
	}
	configData, err := json.Marshal(specValue)
	if err != nil {
		t.Fatal(err)
	}
	configPath := path.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatal(err)
	}
	stateData, err := json.Marshal(spec.State{Bundle: tempDir})
	if err != nil {
		t.Fatal(err)
	}
	resultSpec := loadSpec(bytes.NewReader(stateData))
	assert.True(t, reflect.DeepEqual(resultSpec, specValue))
}

func Test_archiveUpperDirs(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	nestedFileData := []byte("MOCK_CONTENT")
	nestedFileDir := path.Join(srcDir, "nested", "dir")
	nestedFilePath := path.Join(nestedFileDir, "file.txt")
	err = os.MkdirAll(nestedFileDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(nestedFilePath, nestedFileData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	destDir, err := os.MkdirTemp("", "dest")
	if err != nil {
		t.Fatal(err)
	}
	containerSpec := spec.Spec{
		Version: spec.Version,
		Mounts: []spec.Mount{
			{
				Destination: "/dev",
				Source:      "tmpfs",
				Type:        "tmpfs",
				Options: []string{
					"nosuid",
					"strictatime",
					"mode=755",
					"size=65536k",
				},
			},
			{
				Destination: "/data",
				Source:      "/path/to/source",
				Type:        "overlay",
				Options: []string{
					"lowerdir=/path/to/lower",
					fmt.Sprintf("upperdir=%s", srcDir),
					"workdir=/path/to/work",
					"private",
				},
			},
		},
	}
	archives := map[string]Archive{"/data": {Src: "/data", Dest: destDir, Name: "data"}}
	archiveUpperDirs(containerSpec, archives)

	destNestedFileDir := path.Join(destDir, "nested", "dir")
	destNestedFilePath := path.Join(destNestedFileDir, "file.txt")
	resultData, err := os.ReadFile(destNestedFilePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(resultData), string(nestedFileData))
	fileInfo, err := os.Stat(destNestedFilePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fileInfo.Mode().Perm(), fs.FileMode(0600))
	fileInfo, err = os.Stat(destNestedFileDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fileInfo.Mode().Perm(), fs.FileMode(0755))
}
