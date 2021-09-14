//go:build db_acc_test
// +build db_acc_test

package mcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDatabaseInstanceDatastore(t *testing.T) {
	datastore := []interface{}{
		map[string]interface{}{
			"version": "foo",
			"type":    "bar",
		},
	}

	expected := dataStore{
		Version: "foo",
		Type:    "bar",
	}

	actual, _ := extractDatabaseInstanceDatastore(datastore)
	assert.Equal(t, expected, actual)
}

func TestExtractDatabaseInstanceNetworks(t *testing.T) {
	network := []interface{}{
		map[string]interface{}{
			"uuid":        "foobar",
			"port":        "",
			"fixed_ip_v4": "",
		},
	}

	expected := []networkOpts{
		{
			UUID: "foobar",
		},
	}

	actual, _ := extractDatabaseInstanceNetworks(network)
	assert.Equal(t, expected, actual)
}

func TestExtractDatabaseInstanceAutoExpand(t *testing.T) {
	autoExpand := []interface{}{
		map[string]interface{}{
			"autoexpand":    true,
			"max_disk_size": 1000,
		},
	}

	expected := instanceAutoExpandOpts{
		AutoExpand:  true,
		MaxDiskSize: 1000,
	}

	actual, _ := extractDatabaseInstanceAutoExpand(autoExpand)
	assert.Equal(t, expected, actual)
}

func TestExtractDatabaseInstanceWalVolume(t *testing.T) {
	walVolume := []interface{}{
		map[string]interface{}{
			"size":          10,
			"volume_type":   "ms1",
			"autoexpand":    true,
			"max_disk_size": 1000,
		},
	}

	expected := walVolumeOpts{
		Size:        10,
		VolumeType:  "ms1",
		AutoExpand:  true,
		MaxDiskSize: 1000,
	}

	actual, _ := extractDatabaseInstanceWalVolume(walVolume)
	assert.Equal(t, expected, actual)
}

func TestExtractDatabaseInstanceCapabilities(t *testing.T) {
	capabilities := []interface{}{
		map[string]interface{}{
			"name": "node_exporter",
			"settings": map[string]string{
				"listen_port": "9100",
			},
		},
		map[string]interface{}{
			"name": "mysqld_exporter",
			"settings": map[string]string{
				"listen_port": "9104",
			},
		},
	}

	expected := []instanceCapabilityOpts{
		{
			Name: "node_exporter",
			Params: map[string]string{
				"listen_port": "9100",
			},
		},
		{
			Name: "mysqld_exporter",
			Params: map[string]string{
				"listen_port": "9104",
			},
		},
	}

	actual, _ := extractDatabaseInstanceCapabilities(capabilities)
	assert.Equal(t, expected, actual)
}
