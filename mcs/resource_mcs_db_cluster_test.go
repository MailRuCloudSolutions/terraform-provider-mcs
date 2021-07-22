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

func testAccCheckDatabaseClusterExists(n string, cluster *dbClusterResp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(Config)
		DatabaseClient, err := config.DatabaseV1Client(OSRegionName)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack compute client: %s", err)
		}

		found, err := dbClusterGet(DatabaseClient, rs.Primary.ID).extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Cluster not found")
		}

		*cluster = *found

		return nil
	}
}

func testAccCheckDatabaseClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(Config)

	DatabaseClient, err := config.DatabaseV1Client(OSRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack database client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_db_cluster" {
			continue
		}
		_, err := dbClusterGet(DatabaseClient, rs.Primary.ID).extract()
		if err == nil {
			return fmt.Errorf("Cluster still exists")
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
`, OSFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID)

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
`, OSNewFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID)
