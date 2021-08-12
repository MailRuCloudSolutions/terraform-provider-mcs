---
layout: "mcs"
page_title: "mcs: region"
description: |-
Get information about region.
---

`mcs_region` provides details about a specific MCS region.

As well as validating a given region name this resource can be used to discover the name of the region configured within the provider.

### Example Usage

The following example shows how the resource might be used to obtain the name of the AWS region configured on the provider.

```hcl
data "mcs_region" "current" {}
```

### Argument Reference

The following arguments are supported:

* `id` - (Optional) ID of the region to learn or use. Use empty value to learn current region on the provider.

### Attributes Reference

* `id` - Id of the region.
* `parent_region` - Parent of the region.
* `description` - Description of the region.

