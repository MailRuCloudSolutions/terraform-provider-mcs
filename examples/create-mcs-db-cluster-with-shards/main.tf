terraform {
  required_providers {
    mcs = {
      source  = "MailRuCloudSolutions/mcs"
      version = "~> 0.3.0"
    }
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

data "openstack_compute_flavor_v2" "db" {
  name = var.db-instance-flavor
}

resource "openstack_networking_network_v2" "db" {
  name           = "db-net"
  admin_state_up = true
}

resource "mcs_db_cluster_with_shards" "db-cluster-with-shards" {
  name = "db-cluster-with-shards"

  datastore {
    type    = "clickhouse"
    version = "20.8"
  }

  shard {
    size        = 2
    shard_id    = "shard0"
    flavor_id   = data.openstack_compute_flavor_v2.db.id

    volume_size = 10
    volume_type = "ceph-ssd"

    network {
      uuid = openstack_networking_network_v2.db.id
    }
  }

  shard {
    size        = 2
    shard_id    = "shard1"
    flavor_id   = data.openstack_compute_flavor_v2.db.id
    
    volume_size = 10
    volume_type = "ceph-ssd"

    network {
      uuid = openstack_networking_network_v2.db.id
    }
  }
}