package gcp

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ianbrown78/gcp-stop/config"
	"github.com/ianbrown78/gcp-stop/helpers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/syncmap"
	"google.golang.org/api/sqladmin/v1"
)

type SqlInstances struct {
	serviceClient *sqladmin.Service
	base          ResourceBase
	resourceMap   syncmap.Map
}

func init() {}

// Name - Name of the resourceLister for SqlInstances
func (c *SqlInstances) Name() string {
	return "SqlInstances"
}

// ToSlice - Name of the resourceLister for SqlInstances
func (c *SqlInstances) ToSlice() (slice []string) {
	return helpers.SortedSyncMapKeys(&c.resourceMap)
}

// Setup - populates the struct
func (c *SqlInstances) Setup(config config.Config) {
	c.base.config = config
}

// List - Returns a list of all SqlInstances
func (c *SqlInstances) List(refreshCache bool) []string {
	if !refreshCache {
		return c.ToSlice()
	}
	c.resourceMap = sync.Map{}

	for _, zone := range c.base.config.Zones {
		instanceListCall := c.serviceClient.Instances.List(c.base.config.Project)
		instanceList, err := instanceListCall.Do()
		if err != nil {
			log.Fatal(err)
		}

		for _, instance := range instanceList.Items {
			instanceResource := DefaultResourceProperties{
				zone: zone,
			}
			c.resourceMap.Store(instance.Name, instanceResource)
		}
	}

	return c.ToSlice()
}

// Dependencies - Returns a List of resource names to check for
func (c *SqlInstances) Dependencies() []string {
	return []string{}
}

// Shutdown instances
func (c *SqlInstances) Shutdown() error {
	errs, _ := errgroup.WithContext(c.base.config.Context)

	c.resourceMap.Range(func(key, value interface{}) bool {
		instanceID := key.(string)
		zone := value.(DefaultResourceProperties).zone

		rb := &sqladmin.DatabaseInstance{
			Settings: &sqladmin.Settings{
				ActivationPolicy: "NEVER",
			},
		}

		// Parallel instance shutdown
		errs.Go(func() error {
			getInstanceCall := c.serviceClient.Instances.Get(c.base.config.Project, instanceID)
			_, err := getInstanceCall.Do()
			if err != nil {
				return err
			}

			shutdownCall := c.serviceClient.Instances.Patch(c.base.config.Project, instanceID, rb)
			operation, err := shutdownCall.Do()
			if err != nil {
				return err
			}

			var opStatus string
			seconds := 0
			for opStatus != "DONE" {
				log.Printf("[Info] Resource currently being shutdown %v [type: %v project: %v zone %v] (%v seconds)",
					instanceID, c.Name(), c.base.config.Project, zone, seconds)

				operationCall := c.serviceClient.Operations.Get(c.base.config.Project, operation.Name)
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
			// Remove the resource from the list to shut down.
			c.resourceMap.Delete(instanceID)

			log.Printf("[Info] Resource shutdown %v [type: %v project: %v zone: %v] (%v seconds)",
				instanceID, c.Name(), c.base.config.Project, zone, seconds)
			return nil
		})
		return true
	})

	err := errs.Wait()
	return err
}
