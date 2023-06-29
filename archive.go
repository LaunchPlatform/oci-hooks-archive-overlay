package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
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
	// The user (uid) to set for the files inside the tar archive
	TarUser int
	// The group (gid) to set for the files inside the tar archive
	TarGroup int
}

const (
	ArchiveMethodCopy    string = "copy"
	ArchiveMethodTarGzip        = "tar.gz"
)

const (
	annotationPrefix             string = "com.launchplatform.oci-hooks.archive-overlay."
	annotationMountPointArg      string = "mount-point"
	annotationArchiveToArg       string = "archive-to"
	annotationMethodArg          string = "method"
	annotationSuccessArg         string = "success"
	annotationTarContentOwnerArg string = "tar-content-owner"
)

func parseOwner(owner string) (int, int, error) {
	parts := strings.Split(owner, ":")
	if len(parts) < 1 || len(parts) > 2 {
		return 0, 0, fmt.Errorf("Expected only one or two parts in the owner but got %d instead", len(parts))
	}
	uid, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	if len(parts) == 1 {
		return uid, 0, nil
	}
	gid, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

func parseArchives(annotations map[string]string) map[string]Archive {
	archives := map[string]Archive{}
	for key, value := range annotations {
		if !strings.HasPrefix(key, annotationPrefix) {
			continue
		}
		keySuffix := key[len(annotationPrefix):]
		parts := strings.Split(keySuffix, ".")
		name, archiveArg := parts[0], parts[1]
		archive, ok := archives[name]
		if !ok {
			archive = Archive{Name: name, TarUser: -1, TarGroup: -1}
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
		case annotationTarContentOwnerArg:
			uid, gid, err := parseOwner(value)
			if err != nil {
				log.Warnf("Invalid owner argument for %s with error %s, ignored", name, err)
				continue
			}
			if uid < 0 || gid < 0 {
				log.Warnf("Invalid owner argument for %s with negative uid or gid, ignored", name)
				continue
			}
			archive.TarUser = uid
			archive.TarGroup = gid
		default:
			log.Warnf("Invalid archive argument %s for archive %s, ignored", archiveArg, name)
			continue
		}
		archives[name] = archive
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
