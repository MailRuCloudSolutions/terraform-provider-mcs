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

resource "openstack_compute_keypair_v2" "keypair" {
  name       = "default"
  public_key = file(var.public-key-file)
}

resource "openstack_networking_network_v2" "db" {
  name           = "db-net"
  admin_state_up = true
}

resource "mcs_db_instance" "db-instance" {
  name        = "db-instance"

  datastore {
    type    = "mysql"
    version = "5.7"
  }

  keypair           = openstack_compute_keypair_v2.keypair.id
  public_access     = true

  flavor_id   = data.openstack_compute_flavor_v2.db.id
  
  size        = 8
  volume_type = "ceph-ssd"
  disk_autoexpand {
    autoexpand    = true
    max_disk_size = 1000
  }

  network {
    uuid = openstack_networking_network_v2.db.id
  }

  capabilities {
    name = "node_exporter"
    settings = {
      "listen_port" : "9100"
    }
  }
}

resource "mcs_db_instance" "db-replica" {
  name        = "db-instance-replica"
  datastore {
    type    = "mysql"
    version = "5.7"
  }
  replica_of  = mcs_db_instance.db-instance.id

  flavor_id   = data.openstack_compute_flavor_v2.db.id

  size        = 8
  volume_type = "ceph-ssd"

  network {
    uuid = openstack_networking_network_v2.db.id
  }
}