package archive_overlay

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	"log"
	"os"
	"path"
	"strings"
)

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

	destArchives := ParseArchives(config.Annotations)

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
