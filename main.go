package main

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"log"
	"os"
	"path"
	"strings"
)

const (
	annotationPrefix string = "com.launchplatform.oci-hooks.archive-overlay."
)

type Archive struct {
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
		archiveKey, archiveArg := parts[0], parts[1]
		archive, ok := archives[archiveKey]
		if !ok {
			archive = Archive{}
		}
		if archiveArg == "src" {
			archive.Src = value
		} else if archiveArg == "dest" {
			archive.Dest = value
		} else {
			log.Printf("Invalid archive argument %s\n", archiveArg)
			continue
		}
		archives[archiveKey] = archive
	}

}
