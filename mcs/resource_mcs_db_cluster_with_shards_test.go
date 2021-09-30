package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabaseClusterWithShards_basic(t *testing.T) {
	var cluster dbClusterResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseClusterWithShardsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseClusterWithShardsBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster_with_shards.basic", &cluster),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_cluster_with_shards.basic", "name", &cluster.Name),
				),
			},
		},
	})
}

func testAccCheckDatabaseClusterWithShardsDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(configer)

	DatabaseClient, err := config.DatabaseV1Client(osRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack database client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_db_cluster_with_shards" {
			continue
		}
		_, err := dbClusterGet(DatabaseClient, rs.Primary.ID).extract()
		if err == nil {
			return fmt.Errorf("cluster still exists")
		}
	}

	return nil
}

var testAccDatabaseClusterWithShardsBasic = fmt.Sprintf(`
 resource "mcs_db_cluster_with_shards" "basic" {
	name      = "basic"

	datastore {
	  version = "%s"
	  type    = "%s"
	}
  
  
	shard {
	  size = 2
	  shard_id = "shard0"
	  flavor_id = "%s"
	  volume_size      = 8
	  volume_type = "ms1"
	  network {
		  uuid = "%s"
	  }
	  availability_zone = "MS1"
	}
  
	shard {
	  size = 1
	  shard_id = "shard1"
	  flavor_id = "%s"
	  volume_size = 8
	  volume_type = "ms1"
	  network {
		   uuid = "%s"
	  }
	  availability_zone = "MS1"
	}
 }
`, osDBShardsDatastoreVersion, osDBShardsDatastoreType, osFlavorID, osNetworkID, osFlavorID, osNetworkID)
