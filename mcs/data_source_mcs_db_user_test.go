package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatabaseDataSourceUser_basic(t *testing.T) {
	resourceName := "mcs_db_user.basic"
	datasourceName := "data.mcs_db_user.basic"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckDatabase(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatabaseUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatabaseUserBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceDatabaseUserID(datasourceName),
					resource.TestCheckResourceAttrPair(resourceName, "name", datasourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceDatabaseUserID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find user data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("User data source ID not set")
		}

		return nil
	}
}

var testAccDataSourceDatabaseUserBasic = fmt.Sprintf(`
%s

data "mcs_db_user" "basic" {
	id = "${mcs_db_user.basic.id}"
}
`, testAccDatabaseUserBasic)
