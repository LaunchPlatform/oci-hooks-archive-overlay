package archive_overlay

import (
	"log"
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
	upperDirPrefix    string = "upperdir="
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
		if archiveArg == annotationSrcArg {
			archive.Src = value
		} else if archiveArg == annotationDestArg {
			archive.Dest = value
		} else {
			log.Printf("Invalid archive argument %s for archive %s\n", archiveArg, archiveName)
			continue
		}
		archives[archiveName] = archive
	}

	// Create map from dest to archives
	destArchives := map[string]Archive{}
	for _, archive := range archives {
		destArchives[archive.Name] = archive
	}
	return destArchives
}
