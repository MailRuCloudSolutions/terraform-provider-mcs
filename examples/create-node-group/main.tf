terraform {
  required_providers {
    mcs = {
      source = "MailRuCloudSolutions/mcs"
      version = "~> 0.5.5"
    }
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

data "mcs_kubernetes_cluster" "your_cluster" {
  cluster_id = "your_cluster_uuid"
}

resource "mcs_kubernetes_node_group" "default_ng" {
  cluster_id = data.mcs_kubernetes_cluster.your_cluster.id

  node_count = 1
  name = "default"
  max_nodes = 5
  min_nodes = 1

  labels {
    key = "env"
    value = "test"
  }

  labels {
    key = "disktype"
    value = "ssd"
  }
  
  taints {
    key = "taintkey1"
    value = "taintvalue1"
    effect = "PreferNoSchedule"
  }

  taints {
    key = "taintkey2"
    value = "taintvalue2"
    effect = "PreferNoSchedule"
  }
}
