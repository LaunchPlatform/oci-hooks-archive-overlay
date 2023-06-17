package main

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	"log"
	"os"
	"path"
	"strings"
)

const (
	annotationPrefix  string = "com.launchplatform.oci-hooks.archive-overlay."
	annotationSrcArg  string = "src"
	annotationDestArg string = "dest"
	upperDirPrefix    string = "upperdir="
)

type Archive struct {
	Name string
	Src  string
	Dest string
}

func main() {
	var state spec.State
	err := json.NewDecoder(os.Stdin).Decode(&state)
	if err != nil {
		log.Fatal(err)
	}
	jsonFile, err := os.Open(path.Join(state.Bundle, "config.json"))
	if err != nil {
		log.Fatal(err)
	}
	var config spec.Spec
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	archives := map[string]Archive{}
	for key, value := range config.Annotations {
		if !strings.HasPrefix(key, annotationPrefix) {
			continue
		}
		keySuffix := key[len(annotationPrefix):]
		parts := strings.Split(keySuffix, ".")
		archiveName, archiveArg := parts[0], parts[1]
		archive, ok := archives[archiveName]
		if !ok {
			archive = Archive{Name: archiveName}
		}
		if archiveArg == annotationSrcArg {
			archive.Src = value
		} else if archiveArg == annotationDestArg {
			archive.Dest = value
		} else {
			log.Printf("Invalid archive argument %s\n", archiveArg)
			continue
		}
		archives[archiveName] = archive
	}

	// Create map from dest to archives
	destArchives := map[string]Archive{}
	for _, archive := range archives {
		destArchives[archive.Name] = archive
	}

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
			log.Fatalf("Cannot find upperdir for archive %s in mount %s\n", archive.Name, string(mountJson))
		}
		err := cp.Copy(upperDir, archive.Dest)
		if err != nil {
			log.Fatalf("Failed to copy from %s to %s for archive %s with error %s", upperDir, archive.Dest, archive.Name, err)
		}
	}
}
