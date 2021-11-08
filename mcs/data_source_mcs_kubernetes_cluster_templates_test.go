package mcs

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccKubernetesDataSourceClusterTemplates(t *testing.T) {
	tests := map[string]struct {
		name     string
		testCase resource.TestCase
	}{
		"no filter": {
			name: "data.mcs_kubernetes_clustertemplates.empty",
			testCase: resource.TestCase{
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testAccDataSourceMCSKubernetesClusterTemplatesConfig(),
						Check: resource.ComposeTestCheckFunc(
							testAccDataSourceMCSKubernetesClusterTemplatesCheck("data.mcs_kubernetes_clustertemplates.empty"),
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

func testAccDataSourceMCSKubernetesClusterTemplatesConfig() string {
	return `
data "mcs_kubernetes_clustertemplates" "empty" {}
`
}

func testAccDataSourceMCSKubernetesClusterTemplatesCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		templates, ok := rs.Primary.Attributes["cluster_templates.#"]
		if !ok {
			return fmt.Errorf("cluster_templates attribute is missing.")
		}

		templatesQuantity, err := strconv.Atoi(templates)
		if err != nil {
			return fmt.Errorf("error parsing templates (%s) into integer: %s", templates, err)
		}

		if templatesQuantity == 0 {
			return fmt.Errorf("No templates found, this is probably a bug.")
		}

		return nil
	}
}
