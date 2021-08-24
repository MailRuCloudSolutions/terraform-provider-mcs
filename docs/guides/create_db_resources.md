---
layout: "mcs"
page_title: "Create Database Resources"
description: |-
  Create Database Resources
---

# Create Database Instance
This example shows how to create a simple database instance from scratch.
First you have to set up basic configuration as shown in [Getting Started](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs/guides/getting_started)

Create database instance:

```hcl
resource "mcs_db_instance" "db-instance" {
  name        = "db-instance"

  datastore {
    type    = "mysql"
    version = "5.7"
  }
  keypair           = openstack_compute_keypair_v2.mykeypair.id
  public_access     = true

  flavor_id   = data.openstack_compute_flavor_v2.myflavor.id
  
  size        = 8
  volume_type = "ceph-ssd"
  disk_autoexpand {
    autoexpand    = true
    max_disk_size = 1000
  }

  network {
    uuid = openstack_networking_network_v2.mynet.id
  }

  capabilities {
    name = "node_exporter"
    settings = {
      "listen_port" : "9100"
    }
  }
}
```

# Create Database Cluster
This example shows how to create a simple database instance from scratch.
First you have to set up basic configuration as shown in [Getting Started](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs/guides/getting_started)

Create database cluster:

```hcl
resource "mcs_db_cluster" "db-cluster" {
  name        = "db-cluster"

  datastore {
    type    = "postgresql"
    version = "12"
  }

  cluster_size = 3

  flavor_id   = data.openstack_compute_flavor_v2.myflavor.id

  volume_size = 10
  volume_type = "ceph-ssd"

  network {
    uuid = openstack_networking_network_v2.mynet.id
  }
}
```
