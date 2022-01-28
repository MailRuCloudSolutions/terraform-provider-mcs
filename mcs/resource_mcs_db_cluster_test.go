package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabaseCluster_basic(t *testing.T) {
	var cluster dbClusterResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseClusterBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster.basic", &cluster),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_cluster.basic", "name", &cluster.Name),
				),
			},
			{
				Config: testAccDatabaseClusterUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster.basic", &cluster),
					resource.TestCheckResourceAttr(
						"mcs_db_cluster.basic", "volume_size", "9"),
				),
			},
		},
	})
}

func TestAccDatabaseCluster_wal(t *testing.T) {
	var cluster dbClusterResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseClusterBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster.basic", &cluster),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_cluster.basic", "name", &cluster.Name),
				),
			},
		},
	})
}

func TestAccDatabaseCluster_wal_no_update(t *testing.T) {
	var cluster dbClusterResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseClusterWal,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster.basic", &cluster),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_cluster.basic", "name", &cluster.Name),
				),
			},
			{
				Config: testAccDatabaseClusterWal,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseClusterExists(
						"mcs_db_cluster.basic", &cluster),
				),
			},
		},
	})
}

func testAccCheckDatabaseClusterExists(n string, cluster *dbClusterResp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no id is set")
		}

		config := testAccProvider.Meta().(configer)
		DatabaseClient, err := config.DatabaseV1Client(osRegionName)
		if err != nil {
			return fmt.Errorf("error creating openstack compute client: %s", err)
		}

		found, err := dbClusterGet(DatabaseClient, rs.Primary.ID).extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("cluster not found")
		}

		*cluster = *found

		return nil
	}
}

func testAccCheckDatabaseClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(configer)

	DatabaseClient, err := config.DatabaseV1Client(osRegionName)
	if err != nil {
		return fmt.Errorf("error creating openstack database client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_db_cluster" {
			continue
		}
		_, err := dbClusterGet(DatabaseClient, rs.Primary.ID).extract()
		if err == nil {
			return fmt.Errorf("cluster still exists")
		}
	}

	return nil
}

var testAccDatabaseClusterBasic = fmt.Sprintf(`
 resource "mcs_db_cluster" "basic" {
   name      = "basic"
   flavor_id = "%s"
   volume_size      = 8
   volume_type = "ms1"
   cluster_size = 3
   datastore {
	version = "%s"
	type    = "%s"
  }

   network {
     uuid = "%s"
   }
	
   availability_zone = "MS1"
 }
`, osFlavorID, osDBDatastoreVersion, osDBDatastoreType, osNetworkID)

var testAccDatabaseClusterUpdate = fmt.Sprintf(`
resource "mcs_db_cluster" "basic" {
	name      = "basic"
	flavor_id = "%s"
	volume_size      = 9
	volume_type = "ms1"
	cluster_size = 3
	datastore {
	 version = "%s"
	 type    = "%s"
   }
 
	network {
	  uuid = "%s"
	}
	 
	availability_zone = "MS1"
  }
`, osNewFlavorID, osDBDatastoreVersion, osDBDatastoreType, osNetworkID)

var testAccDatabaseClusterWal = fmt.Sprintf(`
 resource "mcs_db_cluster" "basic" {
   name      = "basic"
   flavor_id = "%s"
   volume_size      = 8
   volume_type = "ms1"
   cluster_size = 3
   datastore {
	version = "%s"
	type    = "%s"
  }

   network {
     uuid = "%s"
   }
	
   availability_zone = "MS1"
   wal_volume {
	size = 8
	volume_type = "ms1"
   }

   wal_disk_autoexpand {
	autoexpand = true
	max_disk_size = 1000
   }
 }
`, osFlavorID, osDBDatastoreVersion, osDBDatastoreType, osNetworkID)
