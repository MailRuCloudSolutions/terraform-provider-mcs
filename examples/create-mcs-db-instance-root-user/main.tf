terraform {
  required_providers {
    mcs = {
      source  = "MailRuCloudSolutions/mcs"
      version = "~> 0.5.7"
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
  name           = "k8s-net"
  admin_state_up = true
}

resource "mcs_db_instance" "db-instance" {
  name        = "db-instance"

  datastore {
    type    = "postgresql"
    version = "10"
  }

  flavor_id   = data.openstack_compute_flavor_v2.db.id
  
  size        = 8
  volume_type = "ceph-ssd"
  network {
    uuid = openstack_networking_network_v2.db.id
  }

  root_enabled  = true
  root_password = var.db-root-user-pwd
}

output "root_user_password" {
  value = mcs_db_instance.db-instance.root_password
}

