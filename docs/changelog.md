---
layout: "mcs"
page_title: "MCS provider changelog"
description: |-
  The MCS provider's changelog.
---

# MCS Provider's changelog

#### v0.5.6
- Make `name` attribute of node group required.

#### v0.5.4
- Added `loadbalancer_subnet_id` attribute to cluster.

#### v0.5.0
- Added `availability_zones` attribute to cluster node group.
- Added `mcs_kubernetes_clustertemplates` data source.

#### v0.4.0
- Added `region` support for provider.
- Added `mcs_region` and `mcs_regions` data sources.

#### v0.3.4
- Removed field `node_count` for kubernetes cluster.

#### v0.3.3
- Added required field `availablity_zone` to kubernetes cluster.