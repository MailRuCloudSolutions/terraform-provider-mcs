package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/db/v1/databases"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func databaseDatabaseStateRefreshFunc(client databaseClient, instanceID string, databaseName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pages, err := databaseList(client, instanceID).AllPages()
		if err != nil {
			return nil, "", fmt.Errorf("unable to retrieve mcs database databases: %s", err)
		}

		allDatabases, err := databases.ExtractDBs(pages)
		if err != nil {
			return nil, "", fmt.Errorf("unable to extract mcs database databases: %s", err)
		}

		for _, v := range allDatabases {
			if v.Name == databaseName {
				return v, "ACTIVE", nil
			}
		}

		return nil, "BUILD", nil
	}
}

func databaseDatabaseExists(client databaseClient, instanceID string, databaseName string) (bool, error) {
	var exists bool
	var err error

	pages, err := databaseList(client, instanceID).AllPages()
	if err != nil {
		return exists, err
	}

	allDatabases, err := databases.ExtractDBs(pages)
	if err != nil {
		return exists, err
	}

	for _, v := range allDatabases {
		if v.Name == databaseName {
			exists = true
			return exists, nil
		}
	}

	return exists, err
}
