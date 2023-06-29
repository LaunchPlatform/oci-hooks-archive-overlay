package main

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type Archive struct {
	// The name of archive
	Name string
	// The "destination" filed of overlay mount point to get upperdir folder from
	MountPoint string
	// The destination for copying the upperdir folder to
	ArchiveTo string
	// The empty file to create for indicating archive is done successfully
	ArchiveSuccess string
	// Archive method
	Method string
}

const (
	ArchiveMethodCopy    string = "copy"
	ArchiveMethodTarGzip        = "tar.gz"
)

const (
	annotationPrefix        string = "com.launchplatform.oci-hooks.archive-overlay."
	annotationMountPointArg string = "mount-point"
	annotationArchiveToArg  string = "archive-to"
	annotationMethodArg     string = "method"
	annotationSuccessArg    string = "success"
)

func parseArchives(annotations map[string]string) map[string]Archive {
	archives := map[string]Archive{}
	for key, value := range annotations {
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
		switch archiveArg {
		case annotationMountPointArg:
			archive.MountPoint = value
		case annotationArchiveToArg:
			archive.ArchiveTo = value
		case annotationSuccessArg:
			archive.ArchiveSuccess = value
		case annotationMethodArg:
			archive.Method = value
		default:
			log.Warnf("Invalid archive argument %s for archive %s, ignored", archiveArg, archiveName)
			continue
		}
		archives[archiveName] = archive
	}

	// Convert map from using name as the key to use mount-point instead
	mountPointArchives := map[string]Archive{}
	for _, archive := range archives {
		var emptyValue = false
		if archive.MountPoint == "" {
			log.Warnf("Empty mount-point archive argument value for archiving %s, ignored", archive.Name)
			emptyValue = true
		}
		if archive.ArchiveTo == "" {
			log.Warnf("Empty archive-to argument value for archive %s, ignored", archive.Name)
			emptyValue = true
		}
		if archive.Method != "" && archive.Method != ArchiveMethodCopy && archive.Method != ArchiveMethodTarGzip {
			log.Warnf("Invalid method argument value %s for archive %s, ignored", archive.Method, archive.Name)
			emptyValue = true
		}
		if emptyValue {
			continue
		}
		mountPointArchives[archive.MountPoint] = archive
	}
	return mountPointArchives
}
