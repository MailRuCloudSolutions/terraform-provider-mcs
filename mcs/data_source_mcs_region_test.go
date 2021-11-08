package mcs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccKubernetesDataSourceMCSRegion(t *testing.T) {
	tests := map[string]struct {
		name     string
		testCase resource.TestCase
	}{
		"no params": {
			name: "data.mcs_region.empty",
			testCase: resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: `data "mcs_region" "empty" {}`,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("data.mcs_region.empty", "id", "RegionOne"),
							resource.TestCheckResourceAttr("data.mcs_region.empty", "description", ""),
							resource.TestCheckResourceAttr("data.mcs_region.empty", "parent_region", ""),
						),
					},
				},
			},
		},
		"id provided": {
			name: "data.mcs_region.id",
			testCase: resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: `data "mcs_region" "id" {
									id="RegionAms"
								}`,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("data.mcs_region.id", "id", "RegionAms"),
							resource.TestCheckResourceAttr("data.mcs_region.id", "description", ""),
							resource.TestCheckResourceAttr("data.mcs_region.id", "parent_region", ""),
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
