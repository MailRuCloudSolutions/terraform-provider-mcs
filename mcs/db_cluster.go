package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func databaseClusterStateRefreshFunc(client databaseClient, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := dbClusterGet(client, clusterID).extract()
		if err != nil {
			if _, ok := err.(gophercloud.Err404er); ok {
				return c, "DELETED", nil
			}
			return nil, "", err
		}

		for _, inst := range c.Instances {
			if inst.Status == "error" {
				return c, inst.Status, fmt.Errorf("there was an error creating the database cluster")
			}
		}

		return c, c.Task.Name, nil
	}
}
