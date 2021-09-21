---
layout: "mcs"
page_title: "mcs: regions"
description: |-
List available mcs regions.
---

`mcs_regions` provides information about MCS regions. Can be used to filter regions by parent region. To get details of each region the data source can be combined with the `mcs_region` data source.

### Example Usage

Enabled MCS Regions:

```hcl
data "mcs_regions" "current" {}
```

To see regions with the known Parent Region `parent_region_id` argument needs to be set.

```hcl
data "mcs_regions" "current" {
  parent_region_id = "RegionOne"
}
```

### Argument Reference

The following arguments are supported:

* `parent_region_id` - (Optional) ID of the parent region. Use empty value to list all the regions.

### Attributes Reference

* `id` - Random identifier of the data source.
* `names` - Names of regions that meets the criteria.
