---
layout: "mcs"
page_title: "mcs: kubernetes_node_group"
sidebar_current: "docs-kubernetes-node-group"
description: |-
  Get information on clusters node group.
---

# mcs\_kubernetes\_cluster

Provides a cluster node group resource. This can be used to create, modify and delete cluster's node group.

## Example Usage
```
resource "mcs_kubernetes_node_group" "mynodegroup" {
    cluster_id = your_cluster_id
    name = my_new_node_group
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The UUID of the existing cluster.
* `name` - (Optional) The name of node group to create.
 If none presented then random name will be generated.
* `flavor_id` - (Optional) The flavor of this node group.
 Changing this will force to create a new node group.
* `node_count` - (Required) The node count for this node group. Should be greater than 0.
 If `autoscaling_enabled` parameter is set, this attribute will be ignored during update.
* `max_nodes` - (Optional) The maximum allowed nodes for this node group.
* `min_nodes` - (Optional) The minimum allowed nodes for this node group. Default to 0 if not set.
* `volume_size` - (Optional) The size in GB for volume to load nodes from.
 Changing this will force to create a new node group.
* `volume_type` - (Optional) The volume type to load nodes from.
 Changing this will force to create a new node group.
* `autoscaling_enabled` - (Optional) Determines whether the autoscaling is enabled.
* `labels` - (Optional) The list of objects representing representing additional 
    properties of the node group. Each object should have attribute "key". 
    Object may also have optional attribute "value".
* `taints` - (Optional) The list of objects representing node group taints. Each
    object should have following attributes: key, value, effect.

    
## Attributes
`id` is set to the ID of the found cluster template. In addition, the following
attributes are exported:

* `name` - The name of the node group.
* `cluster_id` - The UUID of cluster that node group belongs.
* `node_count` - The count of nodes in node group.
* `max_nodes` - The maximum amount of nodes in node group.
* `min_nodes` - The minimum amount of nodes in node group.
* `volume_size` - The size in GB for volume to load nodes from.
* `volume_type` - The volume type to load nodes from.
* `flavor_id` - The id of flavor.
* `autoscaling_enabled` - Determines whether the autoscaling is enabled.
* `uuid` - The UUID of the cluster's node group.
* `state` - Determines current state of node group (RUNNING, SHUTOFF, ERROR).
* `nodes` - The list of node group's node objects.
* `labels` - The list of key value pairs representing additional
    properties of the node group.
* `taints` - The list of objects representing node group taints.

## Import

Node groups can be imported using the `id`, e.g.

```
$ terraform import mcs_kubernetes_node_group.ng ng_uuid
```