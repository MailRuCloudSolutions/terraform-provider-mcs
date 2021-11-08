package mcs

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccKubernetesDataSourceMCSRegions(t *testing.T) {
	tests := map[string]struct {
		name     string
		testCase resource.TestCase
	}{
		"no filter": {
			name: "data.mcs_regions.empty",
			testCase: resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testAccDataSourceMCSRegionsConfigEmpty(),
						Check: resource.ComposeTestCheckFunc(
							testAccDataSourceMCSRegionsCheck("data.mcs_regions.empty"),
						),
					},
				},
			},
		},
		"with parent id": {
			name: "data.mcs_regions.parent_id",
			testCase: resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testAccDataSourceMCSRegionsConfigParentID(),
						Check: resource.ComposeTestCheckFunc(
							testAccDataSourceMCSRegionsCheck("data.mcs_regions.parent_id"),
						),
					},
				},
			},
		},
	}

	for name := range tests {
		tt := tests[name]
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, tt.testCase)
		})
	}
}

func testAccDataSourceMCSRegionsConfigEmpty() string {
	return `
data "mcs_regions" "empty" {}
`
}

func testAccDataSourceMCSRegionsConfigParentID() string {
	return `
data "mcs_regions" "parent_id" {
	parent_region_id=""
}
`
}

func testAccDataSourceMCSRegionsCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		names, namesOk := rs.Primary.Attributes["names.#"]

		if !namesOk {
			return fmt.Errorf("names attribute is missing.")
		}

		namesQuantity, err := strconv.Atoi(names)

		if err != nil {
			return fmt.Errorf("error parsing names (%s) into integer: %s", names, err)
		}

		if namesQuantity == 0 {
			return fmt.Errorf("No names found, this is probably a bug.")
		}

		return nil
	}
}
