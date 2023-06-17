package main

import (
	"bytes"
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
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
