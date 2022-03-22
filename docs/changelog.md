---
layout: "mcs"
page_title: "MCS provider changelog"
description: |-
  The MCS provider's changelog.
---

# MCS Provider's changelog

#### v0.6.9
- Added support tarantool datastore

#### v0.6.8
- Added `deprecated_at` to `resource_mcs_kubernetes_cluster_template`.

#### v0.6.7
- Removed `node_addresses` and `create_timeout` fields for `resource_mcs_kubernetes_cluster` and `data_mcs_kubernetes_cluster`.

#### v0.6.6
- Added validation for overlapping between cluster and pods subnets

#### v0.6.5
- Fixed update of `mcs_db_user` password
- Fixed crash when wal_volume is absent from configuration but remains in state
- Fixed creation of `mcs_db_instance` without wal_volume for certain datastores
- Disallowed setting of root_enabled and root_password fields for replicas of `mcs_db_instance`

#### v0.6.4
- Fixed importing for resources with wal_volume
- Added wait until capability has been applied for `mcs_db_instance`, `mcs_db_cluster` and `mcs_db_cluster_with_shards`.
- Added validation of dbms existence for `mcs_db_database` and `mcs_db_user`.

#### v0.6.3
- Removed availability zone validation for `mcs_kubernetes_cluster`.

#### v0.6.2
- Fixed `mcs_kubernetes_cluster` read.

#### v0.6.1
- Fixed import of `volume_type` parameter for `mcs_db_instance` and `mcs_db_cluster`.

#### v0.6.0
- Added import functionality for dbaas resources.
- Fixed `fixed_ip_v4` network parameter for `mcs_db_instance`.

#### v0.5.8
- Removed attribute `ingress_floating_ip` from `mcs_kubernetes_cluster`.

#### v0.5.7
- Forbade using name in `master_flavor` attribute in `mcs_kubernetes_cluster`.
- Forbade using name in `flavor_id` attribute in `mcs_kubernetes_nodegroup`.

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
