---
layout: "mcs"
page_title: "Create Kubernetes Cluster"
description: |-
  Create Kubernetes Cluster
---

# Create Kubernetes Cluster
This example shows how to create a simple kubernetes cluster from scratch.
First you have to set up basic configuration as shown in [Getting Started](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs/guides/getting_started)

Create kubernetes cluster:  

```hcl
resource "mcs_kubernetes_cluster_v1" "k8s-cluster" {
  depends_on = [
    openstack_networking_router_interface_v2.k8s,
  ]

  name = "k8s-cluster"
  cluster_template_id = var.k8s-template.id
  master_flavor       = var.k8s-flavor.id
  master_count        = 3

  keypair = openstack_compute_keypair_v2.mykeypair.id
  network_id = openstack_networking_network_v2.mynet.id
  subnet_id = openstack_networking_subnet_v2.mysubnet.id
  floating_ip_enabled = true
}
```
