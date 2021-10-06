package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func extractBlockStorageMetadataMap(v map[string]interface{}) (map[string]string, error) {
	m := make(map[string]string)
	for key, val := range v {
		metadataValue, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("metadata %s value should be string", key)
		}
		m[key] = metadataValue
	}
	return m, nil
}

func blockStorageVolumeStateRefreshFunc(client databaseClient, volumeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := volumeGet(client.(*gophercloud.ServiceClient), volumeID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return v, bsVolumeStatusDeleted, nil
			}
			return nil, "", err
		}
		if v.Status == "error" {
			return v, v.Status, fmt.Errorf("there was an error creating the block storage volume")
		}

		return v, v.Status, nil
	}
}

func blockStorageSnapshotStateRefreshFunc(client databaseClient, volumeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := snapshotGet(client.(*gophercloud.ServiceClient), volumeID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); ok {
				return v, bsSnapshotStatusDeleted, nil
			}
			return nil, "", err
		}
		if v.Status == "error" {
			return v, v.Status, fmt.Errorf("there was an error creating the block storage volume")
		}

		return v, v.Status, nil
	}
}
