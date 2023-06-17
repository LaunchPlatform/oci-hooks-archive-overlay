package main

import (
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"strings"
)

const (
	upperDirPrefix  = "upperdir="
	defaultLogLevel = "info"
)

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logLevel  = defaultLogLevel
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

func archiveUpperDirs(containerSpec spec.Spec, mountPointArchives map[string]Archive) {
	for _, mount := range containerSpec.Mounts {
		if mount.Type != "overlay" {
			log.Warnf("Unexpected mount type %s, only overlap supported for now, ignored mount at %s", mount.Type, mount.Destination)
			continue
		}
		archive, ok := mountPointArchives[mount.Destination]
		if !ok {
			log.Tracef("Cannot find mount point %s to archive, skip", mount.Destination)
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
		log.Infof("Copying upperdir from %s to %s for archive %s", upperDir, archive.ArchiveTo, archive.Name)
		err := cp.Copy(upperDir, archive.ArchiveTo)
		if err != nil {
			log.Fatalf("Failed to copy from %s to %s for archive %s with error %s", upperDir, archive.ArchiveTo, archive.Name, err)
		}
	}
}

func run() {
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

func setupLogLevel() {
	var found = false
	for _, level := range LogLevels {
		if level == strings.ToLower(logLevel) {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Log Level %q is not supported, choose from: %s\n", logLevel, strings.Join(LogLevels, ", "))
		os.Exit(1)
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	log.SetLevel(level)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "archive_overlay [options]",
		Short: "Invoked as a poststop OCI-hooks to archive upperdir of specific overlay mount",
		Run: func(cmd *cobra.Command, args []string) {
			setupLogLevel()
			run()
		},
	}
	pFlags := rootCmd.PersistentFlags()
	logLevelFlagName := "log-level"
	pFlags.StringVar(
		&logLevel,
		logLevelFlagName,
		logLevel,
		fmt.Sprintf("Log messages above specified level (%s)", strings.Join(LogLevels, ", ")),
	)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
