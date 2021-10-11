package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Datastore names
const (
	Redis       = "redis"
	MongoDB     = "mongodb"
	PostgresPro = "postgrespro"
	Galera      = "galera_mysql"
	Postgres    = "postgresql"
	Clickhouse  = "clickhouse"
	MySQL       = "mysql"
)

func getClusterDatastores() []string {
	return []string{Galera, Postgres}
}

func getClusterWithShardsDatastores() []string {
	return []string{Clickhouse}
}

func extractDatabaseInstanceDatastore(v []interface{}) (dataStore, error) {
	var D dataStore
	in := v[0].(map[string]interface{})
	err := mapStructureDecoder(&D, &in, decoderConfig)
	if err != nil {
		return D, err
	}
	return D, nil
}

func extractDatabaseInstanceNetworks(v []interface{}) ([]networkOpts, error) {
	Networks := make([]networkOpts, len(v))
	for i, network := range v {
		var N networkOpts
		err := mapstructure.Decode(network.(map[string]interface{}), &N)
		if err != nil {
			return nil, err
		}
		Networks[i] = N
	}
	return Networks, nil
}

func extractDatabaseInstanceAutoExpand(v []interface{}) (instanceAutoExpandOpts, error) {
	var A instanceAutoExpandOpts
	in := v[0].(map[string]interface{})
	err := mapstructure.Decode(in, &A)
	if err != nil {
		return A, err
	}
	return A, nil
}

func extractDatabaseInstanceWalVolume(v []interface{}) (walVolumeOpts, error) {
	var W walVolumeOpts
	in := v[0].(map[string]interface{})
	err := mapstructure.Decode(in, &W)
	if err != nil {
		return W, err
	}
	return W, nil
}

func extractDatabaseInstanceCapabilities(v []interface{}) ([]instanceCapabilityOpts, error) {
	capabilities := make([]instanceCapabilityOpts, len(v))
	for i, capability := range v {
		var C instanceCapabilityOpts
		err := mapstructure.Decode(capability.(map[string]interface{}), &C)
		if err != nil {
			return nil, err
		}
		capabilities[i] = C
	}
	return capabilities, nil
}

func databaseInstanceStateRefreshFunc(client databaseClient, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		i, err := instanceGet(client, instanceID).extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return i, "DELETED", nil
			}
			return nil, "", err
		}

		if i.Status == "error" {
			return i, i.Status, fmt.Errorf("there was an error creating the database instance")
		}

		return i, i.Status, nil
	}
}
