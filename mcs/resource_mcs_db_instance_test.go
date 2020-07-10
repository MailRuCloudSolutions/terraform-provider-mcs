package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabaseInstance_basic(t *testing.T) {
	var instance instanceResp

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseInstanceBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseInstanceExists(
						"mcs_db_instance.basic", &instance),
					resource.TestCheckResourceAttrPtr(
						"mcs_db_instance.basic", "name", &instance.Name),
				),
			},
			{
				Config: testAccDatabaseInstanceUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseInstanceExists(
						"mcs_db_instance.basic", &instance),
					resource.TestCheckResourceAttr(
						"mcs_db_instance.basic", "size", "9"),
				),
			},
		},
	})
}

func TestAccDatabaseInstance_rootUser(t *testing.T) {
	var instance instanceResp
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseInstanceRootUser,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseInstanceExists(
						"mcs_db_instance.basic", &instance),
					testAccCheckDatabaseRootUserExists(
						"mcs_db_instance.basic", &instance),
				),
			},
		},
	})
}

func testAccCheckDatabaseInstanceExists(n string, instance *instanceResp) resource.TestCheckFunc {
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

		found, err := instanceGet(DatabaseClient, rs.Primary.ID).extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}

		*instance = *found

		return nil
	}
}

func testAccCheckDatabaseInstanceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(Config)

	DatabaseClient, err := config.DatabaseV1Client(OSRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack database client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_db_instance" {
			continue
		}
		_, err := instanceGet(DatabaseClient, rs.Primary.ID).extract()
		if err == nil {
			return fmt.Errorf("Instance still exists")
		}
	}

	return nil
}

func testAccCheckDatabaseRootUserExists(n string, instance *instanceResp) resource.TestCheckFunc {

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
			return fmt.Errorf("Error creating cloud database client: %s", err)
		}

		isRootEnabledResult := instanceRootUserGet(DatabaseClient, rs.Primary.ID)
		isRootEnabled, err := isRootEnabledResult.extract()
		if err != nil {
			return fmt.Errorf("Error checking if root user is enabled for instance: %s: %s", rs.Primary.ID, err)
		}

		if isRootEnabled {
			return nil
		}

		return fmt.Errorf("Root user %s does not exist", n)
	}
}

var testAccDatabaseInstanceBasic = fmt.Sprintf(`
resource "mcs_db_instance" "basic" {
  name             = "basic"
  flavor_id = "%s"
  size = 8
  volume_type = "ms1"

  datastore {
    version = "%s"
    type    = "%s"
  }

  network {
    uuid = "%s"
  }
  availability_zone = "MS1"
  floating_ip_enabled = true
  keypair = "%s"

  disk_autoexpand {
    autoexpand = true
    max_disk_size = 1000
  }

}
`, OSFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID, OSKeypairName)

var testAccDatabaseInstanceUpdate = fmt.Sprintf(`
resource "mcs_db_instance" "basic" {
  name             = "basic"
  flavor_id = "%s"
  size = 9
  volume_type = "ms1"

  datastore {
    version = "%s"
    type    = "%s"
  }

  network {
    uuid = "%s"
  }
  availability_zone = "MS1"
  floating_ip_enabled = true
  keypair = "%s"

  disk_autoexpand {
    autoexpand = true
    max_disk_size = 2000
  }

}
`, OSNewFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID, OSKeypairName)

var testAccDatabaseInstanceRootUser = fmt.Sprintf(`
resource "mcs_db_instance" "basic" {
  name = "basic"
  flavor_id = "%s"
  size = 10
  volume_type = "ms1"

  datastore {
    version = "%s"
    type    = "%s"
  }

  network {
    uuid = "%s"
  }
  root_enabled = true
}
`, OSFlavorID, OSDBDatastoreVersion, OSDBDatastoreType, OSNetworkID)
