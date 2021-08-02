---
layout: "mcs"
page_title: "MCS: create_kubernetes_cluster"
description: |-
  Creating Kubernetes Cluster with Terraform.
---

# Create Kubernetes Cluster with Terraform

This example shows how to create a simple kubernetes cluster from scratch.

First, create a Terraform config file named `main.tf`. Inside, you'll want to include the configuration of
[MCS Provider](/terraform/docs/providers/mcs/index.html),
[Openstack Provider](https://www.terraform.io/docs/providers/openstack/index.html),
[openstack_networking_network_v2](https://www.terraform.io/docs/providers/openstack/d/networking_network_v2.html),
[openstack_networking_subnet_v2](https://www.terraform.io/docs/providers/openstack/r/networking_subnet_v2.html),
[openstack_networking_network_v2](https://www.terraform.io/docs/providers/openstack/d/networking_network_v2.html),
[openstack_networking_router_v2](https://www.terraform.io/docs/providers/openstack/r/networking_router_v2.html),
[openstack_networking_router_interface_v2](https://www.terraform.io/docs/providers/openstack/r/networking_router_interface_v2.html),
[openstack_compute_keypair_v2](https://www.terraform.io/docs/providers/openstack/d/compute_keypair_v2.html)
[openstack_compute_flavor_v2](https://www.terraform.io/docs/providers/openstack/d/compute_flavor_v2.html)
and [kubernetes_cluster](/terraform/docs/providers/mcs/r/mcs_kubernetes_cluster.html).

Use MCS provider:

```hcl
provider "mcs" {
    username   = "some_user"
    password   = "s3cr3t"
    project_id = "some_project_id"
    auth_url   = "https://infra.mail.ru/identity/v3/"
  }
}
```

Configure MCS provider:

* The `username` field should be replaced with your user_name.
* The `password` field should be replaced with your user's password.
* The `project_id` field should be replaced with your project_id.

Use Openstack provider:

```hcl
provider "openstack" {
    user_name        = "your USER_NAME"
    password         = "your PASSWORD"
    auth_url         = "https://infra.mail.ru/identity/v3/"
    tenant_id        =  "your PROJECT_ID"
    user_domain_name = "users"
}
```
**NOTE:** You should not use `OS_USER_DOMAIN_ID` env variable when working with two providers.

Configure Openstack provider:

* The `user_name` field should be replaced with your user_name.
* The `password` field should be replaced with your user's password.
* The `tenant_id` field should be replaced with your project_id.

Create network:
```hcl
resource "openstack_networking_network_v2" "mynet" {
  name           = "mynet"
}
```

Create subnet:

```hcl
resource "openstack_networking_subnet_v2" "mysubnet" {
  name            = "mysubnet"
  network_id      = openstack_networking_network_v2.mynet.id
  cidr            = "10.100.0.0/16"
  ip_version      = 4
}
```

Use external network to get floating IPs:

```hcl
data "openstack_networking_network_v2" "extnet" {
  name = "ext-net"
}
```

Create router connected to external network:

```hcl
resource "openstack_networking_router_v2" "myrouter" {
  name                = "myrouter"
  external_network_id = data.openstack_networking_network_v2.extnet.id
}
```

Connect private network to router

```hcl
resource "openstack_networking_router_interface_v2" "myrouterinterface" {
  router_id = openstack_networking_router_v2.myrouter.id
  subnet_id = openstack_networking_subnet_v2.mysubnet.id
}
```

Use keypair:

```hcl
data "openstack_compute_keypair_v2" "mykeypair" {
  name       = "mykey"
}
```

Use flavor:

```hcl
data "openstack_compute_flavor_v2" "myflavor" {
  name = "Standard-2-4-50"
}
```

Create kubernetes cluster:

```hcl
resource "mcs_kubernetes_cluster_v1" "k8s-cluster" {
  depends_on = [
    openstack_networking_router_interface_v2.k8s,
  ]

  name = "k8s-cluster"
  cluster_template_id = var.k8s-template.id
  flavor              = var.k8s-flavor.id
  master_flavor       = var.k8s-flavor.id
  master_count        = 3

  keypair = openstack_compute_keypair_v2.mykeypair.id
  network_id = openstack_networking_network_v2.mynet.id
  subnet_id = openstack_networking_subnet_v2.mysubnet.id
  floating_ip_enabled = true
}
```
