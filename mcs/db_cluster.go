package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func flattenDatabaseClusterWalVolume(w walVolume) []map[string]interface{} {
	walvolume := make([]map[string]interface{}, 1)
	walvolume[0] = make(map[string]interface{})
	walvolume[0]["size"] = w.Size
	walvolume[0]["volume_type"] = dbImportedStatus
	return walvolume
}

func flattenDatabaseClusterShard(inst dbClusterInstanceResp) map[string]interface{} {
	newShard := make(map[string]interface{})
	newShard["shard_id"] = inst.ShardID
	newShard["flavor_id"] = inst.Flavor.ID
	newShard["volume_size"] = inst.Volume.Size
	newShard["volume_type"] = dbImportedStatus
	return newShard
}

func getClusterStatus(c *dbClusterResp) string {
	instancesStatus := string(dbInstanceStatusActive)
	for _, inst := range c.Instances {
		if inst.Status == "error" {
			return inst.Status
		}
		if inst.Status == string(dbInstanceStatusBuild) || inst.Status == string(dbInstanceStatusResize) {
			instancesStatus = inst.Status
		}
	}
	if c.Task.Name == "NONE" {
		switch instancesStatus {
		case string(dbInstanceStatusActive):
			return string(dbClusterStatusActive)
		case string(dbInstanceStatusBuild):
			return string(dbClusterStatusBuild)
		case string(dbInstanceStatusResize):
			return string(dbClusterStatusResize)
		}
	}

	return c.Task.Name
}

func databaseClusterStateRefreshFunc(client databaseClient, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := dbClusterGet(client, clusterID).extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return c, "DELETED", nil
			}
			return nil, "", err
		}

		clusterStatus := getClusterStatus(c)
		if clusterStatus == "error" {
			return c, clusterStatus, fmt.Errorf("there was an error creating the database cluster")
		}

		return c, clusterStatus, nil
	}
}
