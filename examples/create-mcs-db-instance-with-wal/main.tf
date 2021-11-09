terraform {
  required_providers {
    mcs = {
      source  = "MailRuCloudSolutions/mcs"
      version = "~> 0.5.5"
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

resource "mcs_db_instance" "db-instance" {
  name        = "db-instance"

  datastore {
    type    = "postgresql"
    version = "11"
  }

  public_access     = true

  flavor_id   = data.openstack_compute_flavor_v2.db.id

  size        = 10
  volume_type = "ceph-ssd"
  disk_autoexpand {
    autoexpand    = true
    max_disk_size = 1000
  }
  wal_volume {
    size          = 10
    volume_type   = "ceph-ssd"
    autoexpand    = true
    max_disk_size = 20
  }

  network {
    uuid = openstack_networking_network_v2.db.id
  }
}
