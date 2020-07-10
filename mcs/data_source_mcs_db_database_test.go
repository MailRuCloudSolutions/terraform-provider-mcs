package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceDatabaseDatabase_basic(t *testing.T) {
	resourceName := "mcs_db_database.basic"
	datasourceName := "data.mcs_db_database.basic"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatabaseDatabaseBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceDatabaseDatabaseID(datasourceName),
					resource.TestCheckResourceAttrPair(resourceName, "name", datasourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceDatabaseDatabaseID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find database data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Database data source ID not set")
		}

		return nil
	}
}

var testAccDataSourceDatabaseDatabaseBasic = fmt.Sprintf(`
%s

data "mcs_db_database" "basic" {
	id = "${mcs_db_database.basic.id}"
}
`, testAccDatabaseDatabaseBasic)
