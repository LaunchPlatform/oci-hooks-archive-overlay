package main

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

const (
	upperDirPrefix string = "upperdir="
)

func main() {
	var state spec.State
	err := json.NewDecoder(os.Stdin).Decode(&state)
	if err != nil {
		log.Fatalf("Failed to parse stdin with error %s", err)
	}
	configPath := path.Join(state.Bundle, "config.json")
	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open OCI spec config %s with error %s", configPath, err)
	}
	var config spec.Spec
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to parse OCI spec config JSON file %s with error %s", configPath, err)
	}

	destArchives := ParseArchives(config.Annotations)
	archivesJson, err := json.Marshal(destArchives)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Parsed archives: %s", string(archivesJson))

	for _, mount := range config.Mounts {
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
	log.Infof("Done")
}
