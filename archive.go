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
}

const (
	annotationPrefix        string = "com.launchplatform.oci-hooks.archive-overlay."
	annotationMountPointArg string = "mount-point"
	annotationArchiveToArg  string = "archive-to"
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
		if archiveArg == annotationMountPointArg {
			archive.MountPoint = value
		} else if archiveArg == annotationArchiveToArg {
			archive.ArchiveTo = value
		} else {
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
		if emptyValue {
			continue
		}
		mountPointArchives[archive.MountPoint] = archive
	}
	return mountPointArchives
}
