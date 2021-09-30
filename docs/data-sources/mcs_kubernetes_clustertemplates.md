---
layout: "mcs"
page_title: "mcs: kubernetes cluster templates"
description: |-
List available mcs kubernetes cluster templates.
---

# MCS Kubernetes Cluster Templates

`mcs_kubernetes_cluster_templates` returns list of available MCS Kubernetes Cluster Templates. 
To get details of each cluster template the data source can be combined with the `mcs_kubernetes_clustertemplate` data source.

**New since version v0.5.0**

### Example Usage

Enabled MCS Kubernetes Cluster Templates:

```hcl
data "mcs_kubernetes_clustertemplates" "templates" {}
```

### Attributes Reference

* `id` - Random identifier of the data source.
* `cluster_templates` - A list of available kubernetes cluster templates.
  * `cluster_template_uuid` - The UUID of the cluster template.
  * `name` - The name of the cluster template.


