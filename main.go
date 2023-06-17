package main

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

const (
	upperDirPrefix string = "upperdir="
)

func loadSpec(stateInput io.Reader) spec.Spec {
	var state spec.State
	err := json.NewDecoder(stateInput).Decode(&state)
	if err != nil {
		log.Fatalf("Failed to parse stdin with error %s", err)
	}
	configPath := path.Join(state.Bundle, "config.json")
	jsonFile, err := os.Open(configPath)
	defer jsonFile.Close()
	if err != nil {
		log.Fatalf("Failed to open OCI spec file %s with error %s", configPath, err)
	}
	var containerSpec spec.Spec
	err = json.NewDecoder(jsonFile).Decode(&containerSpec)
	if err != nil {
		log.Fatalf("Failed to parse OCI spec JSON file %s with error %s", configPath, err)
	}
	return containerSpec
}

func archiveUpperDirs(containerSpec spec.Spec, destArchives map[string]Archive) {
	for _, mount := range containerSpec.Mounts {
		archive, ok := destArchives[mount.Destination]
		if !ok {
			continue
		}
		var upperDir = ""
		for _, option := range mount.Options {
			if strings.HasPrefix(option, upperDirPrefix) {
				upperDir = option[len(upperDirPrefix):]
				break
			}
		}
		if upperDir == "" {
			mountJson, err := json.Marshal(mount)
			if err != nil {
				log.Fatal(err)
			}
			log.Fatalf("Cannot find upperdir for archive %s in mount %s", archive.Name, string(mountJson))
		}
		log.Infof("Copying upperdir from %s to %s for archive %s", upperDir, archive.Dest, archive.Name)
		err := cp.Copy(upperDir, archive.Dest)
		if err != nil {
			log.Fatalf("Failed to copy from %s to %s for archive %s with error %s", upperDir, archive.Dest, archive.Name, err)
		}
	}
}

func main() {
	containerSpec := loadSpec(os.Stdin)
	destArchives := parseArchives(containerSpec.Annotations)
	archivesJson, err := json.Marshal(destArchives)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Parsed archives: %s", string(archivesJson))
	archiveUpperDirs(containerSpec, destArchives)
	log.Infof("Done")
}
