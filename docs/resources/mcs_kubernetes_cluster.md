---
layout: "mcs"
page_title: "mcs: kubernetes_cluster"
sidebar_current: "docs-kubernetes-cluster"
description: |-
  Manages a kubernetes cluster.
---

# mcs\_kubernetes\_cluster

Provides a kubernetes cluster resource. This can be used to create, modify and delete kubernetes clusters.

## Example Usage

```terraform

resource "mcs_kubernetes_cluster_v1" "mycluster" {
      name                = "terracluster"
      cluster_template_id = example_template_id
      master_flavor       = example_flavor_id
      master_count        = 1
      network_id          = example_network_id
      subnet_id           = example_subnet_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster. Changing this creates a new cluster.

* `cluster_template_id` - (Required) The UUID of the Kubernetes cluster
    template. It can be obtained using the cluster_template data source

* `master_flavor` - (Optional) The ID of flavor for the master nodes.
 If master_flavor is not present, value from cluster_template will be used

* `network_id` - (Required) The UUID of the network that will be attached to the cluster.
 Changing this creates a new cluster.

* `subnet_id` - (Required) The UUID of the subnet that will be attached to the cluster.
 Changing this creates a new cluster.

* `keypair` - (Optional) The name of the Compute service SSH keypair. Changing
    this creates a new cluster.

* `labels` - (Optional) The list of key value pairs representing additional
    properties of the cluster. Changing this creates a new cluster.

* `master_count` - (Optional) The number of master nodes for the cluster.
    Changing this creates a new cluster.
    
* `pods_network_cidr` - (Optional) The network cidr used in k8s virtual network.

* `floating_ip_enabled` - (Required) Floating ip is enabled.

* `api_lb_vip` - (Optional) API LoadBalancer vip.

* `api_lb_fip` - (Optional) API LoadBalancer fip.

* `ingress_floating_ip` - (Optional) Floating IP created for ingress service.

* `registry_auth_password` - (Optional) Docker registry access password.

## Attributes

This resource exports the following attributes:

* `name` - The name of the cluster.
* `project_id` - The project of the cluster.
* `created_at` - The time at which cluster was created.
* `updated_at` - The time at which cluster was created.
* `api_address` - COE API address.
* `cluster_template_id` - The UUID of the V1 Container Infra cluster template.
* `create_timeout` - The timeout (in minutes) for creating the cluster.
* `discovery_url` - The URL used for cluster node discovery.
* `master_flavor` - The ID of flavor for the master nodes.
* `keypair` - The name of the Compute service SSH keypair.
* `labels` - The list of key value pairs representing additional properties of
                 the cluster.
* `master_count` - The number of master nodes for the cluster.
* `master_addresses` - IP addresses of the master node of the cluster.
* `node_addresses` - IP addresses of the node of the cluster.
* `stack_id` - UUID of the Orchestration service stack.
* `network_id` - UUID of the cluster's network.
* `subnet_id` - UUID of the cluster's subnet.
* `status` - Current state of a cluster. Changing this to `RUNNING` or `SHUTOFF` will turn cluster on/off.
* `pods_network_cidr` - Network cidr of k8s virtual network
* `floating_ip_enabled` - Floating ip is enabled.
* `api_lb_vip` - API LoadBalancer vip.
* `api_lb_fip` - API LoadBalancer fip.
* `ingress_floating_ip` - Floating IP created for ingress service.
* `registry_auth_password` - Docker registry access password.

## Import

Clusters can be imported using the `id`, e.g.

```
$ terraform import mcs_kubernetes_cluster.mycluster ce0f9463-dd25-474b-9fe8-94de63e5e42b
```