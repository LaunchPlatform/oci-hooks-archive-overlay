package main

import (
	"bytes"
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"os"
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {
	cmd := exec.Command(os.Args[0])
	state, err := json.Marshal(spec.State{
		Version: spec.Version,
		ID:      "MOCK_ID",
		Status:  "stopped",
		Bundle:  "/path/to/bundle",
	})
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdin = bytes.NewReader(state)
	err = cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process err %v, want exit status 1", err)
}
