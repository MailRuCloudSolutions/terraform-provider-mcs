package mcs

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccKubernetesNodeGroupDataSource_basic(t *testing.T) {
	var cluster cluster
	var nodeGroup nodeGroup

	clusterName := "testcluster" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	createClusterFixture := clusterFixture(clusterName, clusterTemplateID, osFlavorID,
		osKeypairName, osNetworkID, osSubnetworkID, "MS1", 1)
	clusterResourceName := "mcs_kubernetes_cluster." + clusterName

	nodeGroupName := "testng" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	nodeGroupFixture := nodeGroupFixture(nodeGroupName, osFlavorID, 1, 3, 1, false)
	nodeGroupResourceName := "mcs_kubernetes_node_group." + nodeGroupName
	nodeGroupDataSourceName := "data.mcs_kubernetes_node_group." + nodeGroupName

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheckKubernetes(t) },
		Providers:                 testAccProviders,
		CheckDestroy:              testAccCheckKubernetesClusterDestroy,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesNodeGroupDataSourceBasic(
					testAccKubernetesNodeGroupBasic(clusterName, testAccKubernetesClusterBasic(createClusterFixture), nodeGroupFixture), nodeGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKubernetesClusterExists(clusterResourceName, &cluster),
					testAccCheckKubernetesNodeGroupExists(nodeGroupResourceName, clusterResourceName, &nodeGroup),
					testAccKubernetesNodeGroupDataSourceID(nodeGroupDataSourceName),
					checkNodeGroupAttrs(nodeGroupDataSourceName, nodeGroupFixture),
				),
			},
		},
	})
}

func testAccKubernetesNodeGroupDataSourceID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ct, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find node group data source: %s", resourceName)
		}

		if ct.Primary.ID == "" {
			return fmt.Errorf("node group data source ID is not set")
		}

		return nil
	}
}

func testAccKubernetesNodeGroupDataSourceBasic(nodeGroupResource, nodeGroupName string) string {

	conf := fmt.Sprintf(`
		%[1]s

		data "mcs_kubernetes_node_group" "%[2]s" {
          uuid = "${mcs_kubernetes_node_group.`+nodeGroupName+`.id}"
		}
		`, nodeGroupResource, nodeGroupName)

	return conf
}
