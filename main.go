package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	logrusSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/cobra"
	"io"
	"log/syslog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	upperDirPrefix  = "upperdir="
	defaultLogLevel = "info"
)

var (
	LogLevels = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logLevel  = defaultLogLevel
	useSyslog = false
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

func archiveTarGzip(src string, archiveTo string, uid int, gid int) error {
	// ref: https://golangdocs.com/tar-gzip-in-golang
	// ref: https://github.com/containers/podman/blob/d09edd2820e25372c63e2a9d16a42b6d258b7f80/pkg/bindings/images/build.go#L633-L791
	// ref: https://gist.github.com/mimoo/25fc9716e0f1353791f5908f94d6e726
	archiveFile, err := os.OpenFile(archiveTo, os.O_CREATE|os.O_RDWR, os.FileMode(0644))
	gzipWriter := gzip.NewWriter(archiveFile)
	tarWriter := tar.NewWriter(gzipWriter)
	defer archiveFile.Close()
	defer gzipWriter.Close()
	defer tarWriter.Close()

	srcPath, err := filepath.Abs(src)
	err = filepath.Walk(src, func(path string, fileInfo os.FileInfo, err error) error {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}
		separator := string(filepath.Separator)
		if absPath == srcPath {
			separator = ""
		}
		header.Name = "./" + filepath.ToSlash(strings.TrimPrefix(absPath, srcPath+separator))
		if absPath != srcPath && fileInfo.IsDir() {
			header.Name += "/"
		}
		if uid >= 0 {
			header.Uid = uid
			header.Uname = ""
		}
		if gid >= 0 {
			header.Gid = gid
			header.Gname = ""
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
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

		var method = archive.Method
		if method == "" {
			method = ArchiveMethodCopy
		}
		if method == ArchiveMethodCopy {
			log.Infof("Copying upperdir from %s to %s for archive %s", upperDir, archive.ArchiveTo, archive.Name)
			err := cp.Copy(upperDir, archive.ArchiveTo)
			if err != nil {
				log.Fatalf("Failed to copy from %s to %s for archive %s with error %s", upperDir, archive.ArchiveTo, archive.Name, err)
			}
		} else if method == ArchiveMethodTarGzip {
			log.Infof("Archiving upperdir from %s to %s for archive %s", upperDir, archive.ArchiveTo, archive.Name)
			err := archiveTarGzip(upperDir, archive.ArchiveTo, -1, -1)
			if err != nil {
				log.Fatalf("Failed to archive tar.gz from %s to %s for archive %s with error %s", upperDir, archive.ArchiveTo, archive.Name, err)
			}
		} else {
			log.Fatalf("Unknown archive method %s", method)
		}
		if archive.ArchiveSuccess != "" {
			err := os.WriteFile(archive.ArchiveSuccess, []byte{}, 0644)
			if err != nil {
				log.Fatalf("Failed to write archive success file %s for archive %s with error %s", archive.ArchiveSuccess, archive.Name, err)
			}
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

func initSyslog() {
	if !useSyslog {
		return
	}
	hook, err := logrusSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable syslog with error %s", err)
		return
	}
	log.AddHook(hook)
}

func init() {
	// Hooks are called before PersistentPreRunE(). These hooks affect global
	// state and are executed after processing the command-line, but before
	// actually running the command.
	cobra.OnInitialize(initSyslog)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "archive_overlay [options]",
		Short:   "Invoked as a poststop OCI-hooks to archive upperdir of specific overlay mount",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			setupLogLevel()
			log.Infof("Run archive_overlay %s", Version)
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

	syslogFlagName := "syslog"
	pFlags.BoolVar(
		&useSyslog,
		syslogFlagName,
		useSyslog,
		fmt.Sprintf("Log messages to syslog"),
	)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
