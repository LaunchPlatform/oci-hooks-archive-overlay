package main

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

type Archive struct {
	Name string
	Src  string
	Dest string
}

const (
	annotationPrefix  string = "com.launchplatform.oci-hooks.archive-overlay."
	annotationSrcArg  string = "src"
	annotationDestArg string = "dest"
)

func ParseArchives(annotations map[string]string) map[string]Archive {
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
		if archiveArg == annotationSrcArg {
			archive.Src = value
		} else if archiveArg == annotationDestArg {
			archive.Dest = value
		} else {
			log.Warnf("Invalid archive argument %s for archive %s, ignored", archiveArg, archiveName)
			continue
		}
		archives[archiveName] = archive
	}

	// Convert map from using name as the key to use dest instead
	destArchives := map[string]Archive{}
	for _, archive := range archives {
		var emptyValue = false
		if archive.Src == "" {
			log.Warnf("Empty src archive argument value for archive %s, ignored", archive.Name)
			emptyValue = true
		}
		if archive.Dest == "" {
			log.Warnf("Empty dest archive argument value for archive %s, ignored", archive.Name)
			emptyValue = true
		}
		if emptyValue {
			continue
		}
		destArchives[archive.Dest] = archive
	}
	return destArchives
}
