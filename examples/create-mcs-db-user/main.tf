terraform {
  required_providers {
    mcs = {
      source  = "MailRuCloudSolutions/mcs"
      version = "~> 0.5.1"
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
    type    = "mysql"
    version = "5.7"
  }

  flavor_id   = data.openstack_compute_flavor_v2.db.id

  size        = 8
  volume_type = "ceph-ssd"  

  network {
    uuid = openstack_networking_network_v2.db.id
  }
}

resource "mcs_db_database" "db-database" {
  name        = "testdb"
  instance_id = mcs_db_instance.db-instance.id
  charset     = "utf8"
  collate     = "utf8_general_ci"
}

resource "mcs_db_user" "db-user" {
  name        = "testuser"
  password    = var.db-user-password

  instance_id = mcs_db_instance.db-instance.id

  databases   = [mcs_db_database.db-database.name]
}
