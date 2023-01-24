package gcp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gcp-stop/config"
	"gcp-stop/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/compute/v1"
)

type ComputeInstances struct {
	serviceClient *compute.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {
	computeService, err := compute.NewService(Ctx)
	if err != nil {
		log.Fatal(err)
	}
	computeResource := ComputeInstances{
		serviceClient: computeService,
	}
	register(&computeResource)
}

// Name - Name of the resourceLister for ComputeInstances
func (c *ComputeInstances) Name() string {
	return "ComputeInstances"
}

// ToSlice - Name of the resourceLister for ComputeInstances
func (c *ComputeInstances) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)
}

// Setup - populates the struct
func (c *ComputeInstances) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all ComputeInstances
func (c *ComputeInstances) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	c.resourceMap = sync.Map{}

	for _, zone := range c.base.config.Zones {
		instanceListCall := c.serviceClient.Instances.List(c.base.config.Project, zone)
		instanceList, err := instanceListCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		for _, instance := range instanceList.Items {
			skipInstance := false
			// Skip any managed by instance groups
			for _, item := range instance.Metadata.Items {
				if item.Key == "created-by" && strings.Contains(*item.Value, "/instanceGroupManagers/") {
					skipInstance = true
				}
			}
			if skipInstance {
				continue
			}

			instanceResource := DefaultResourceProperties{
				zone: zone,
			}
			c.resourceMap.Store(instance.Name, instanceResource)
		}
	}
	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *ComputeInstances) Dependencies() []string {
	return []string{}
}

// Shutdown instances
func (c *ComputeInstances) Shutdown() error {
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		instanceID := key.(string)
		zone := value.(DefaultResourceProperties).zone

		// Parallwl instance shutdown
		errs.Go(func() error {
			getInstanceCall := c.serviceClient.Instances.Get(c.base.config.Project, zone, instanceID)
			_, err := getInstanceCall.Do()
			if err != nil {
				return err
			}
			shutdownCall := c.serviceClient.Instances.Stop(c.base.config.Project, zone, instanceID)
			operation, err := shutdownCall.Do()
			if err != nil {
				return err
			}
			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being shutdown %v [type: %v project: %v zone %v] (%v seconds)",
					instanceID, c.Name(), c.base.config.Project, zone, seconds)

				operationCall := c.serviceClient.ZoneOperations.Get(c.base.config.Project, zone, operation.Name)
				checkOp, err := operationCall.Do()
				if err != nil {
					return nil
				}
				opStatus = checkOp.Status

				time.Sleep(time.Duration(c.base.config.PollTime) * time.Second)
				seconds += c.base.config.PollTime
				if seconds > c.base.config.Timeout {
					return fmt.Errorf("[Error] Resource shutdown timed out for %v [type: %v project: %v zone: %v] (%v seconds)",
						instanceID, c.Name(), c.base.config.Project, zone, c.base.config.Timeout)
				}
			}
			c.resourceMap.Delete(instanceID)

			log.Printf("[Info] Resource shutdown %v [type: %v project: %v zone: %v] (%v seconds)",
				instanceID, c.Name(), c.base.config.Project, zone, seconds)
			return nil
		})

		return true
	})

	// Wait for shutdowns to complete, and return the first non nil error
	err := errs.Wait()
	return err
}
