---
layout: "mcs"
page_title: "Getting Started with the MCS Provider"
description: |-
  Getting Started with the MCS Provider
---

# Create basic config for MCS Provider resources

This example shows how to create a simple terraform configuration for creation of MCS resources.

First, create a Terraform config file named `main.tf`. Inside, you'll want to include the configuration of
[MCS Provider](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs),
[Openstack Provider](https://www.terraform.io/docs/providers/openstack/index.html),
[openstack_networking_network_v2](https://www.terraform.io/docs/providers/openstack/d/networking_network_v2.html),
[openstack_networking_subnet_v2](https://www.terraform.io/docs/providers/openstack/r/networking_subnet_v2.html),
[openstack_networking_network_v2](https://www.terraform.io/docs/providers/openstack/d/networking_network_v2.html),
[openstack_networking_router_v2](https://www.terraform.io/docs/providers/openstack/r/networking_router_v2.html),
[openstack_networking_router_interface_v2](https://www.terraform.io/docs/providers/openstack/r/networking_router_interface_v2.html),
[openstack_compute_keypair_v2](https://www.terraform.io/docs/providers/openstack/d/compute_keypair_v2.html)
and [openstack_compute_flavor_v2](https://www.terraform.io/docs/providers/openstack/d/compute_flavor_v2.html).

Use MCS provider:

```hcl
provider "mcs" {
    username   = "some_user"
    password   = "s3cr3t"
    project_id = "some_project_id"
  }
}
```

Configure MCS provider:

* The `username` field should be replaced with your user_name.
* The `password` field should be replaced with your user's password.
* The `project_id` field should be replaced with your project_id.

For additional configuration parameters, please read [configuration reference](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs#configuration-reference)

Use Openstack provider:

```hcl
provider "openstack" {
    user_name        = "your USER_NAME"
    password         = "your PASSWORD"
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

> ⚠️ From v0.5.5 the provider supports **only UUID** of a flavor

```hcl
data "openstack_compute_flavor_v2" "myflavor" {
  name = "b7d20f15-82f1-4ed4-a12e-e60277fe955f" # Standard 2-4-50
}
```
 
You can see the list of all available flavours with the command:

```
openstack flavor list
```
