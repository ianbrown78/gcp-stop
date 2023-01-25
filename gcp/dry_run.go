package gcp

import (
	"log"

	"github.com/ianbrown78/gcp-stop/config"
)

func parallelDryRun(resourceMap map[string]Resource, resource Resource, config config.Config) {
	resourceList := resource.List(false)
	if len(resourceList) == 0 {
		log.Printf("[Dryrun] [Skip] Resource type %v has nothing to shutdown [project: %v]", resource.Name(), config.Project)
		return
	}
	log.Printf("[Dryrun] Resource type %v with resources %v would be shutdown [project: %v]", resource.Name(), resourceList, config.Project)
}
