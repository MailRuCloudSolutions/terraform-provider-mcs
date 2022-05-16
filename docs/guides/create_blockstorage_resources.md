---
layout: "mcs"
page_title: "Create Blockstorage Resources"
description: |-
  Create Blockstorage Resources
---

# Create Blockstorage Volume
This example shows how to create a blockstorage volume from scratch.
First you have to set up basic configuration as shown in [Getting Started](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs/guides/getting_started)

Create blockstorage volume:

```hcl
resource "mcs_blockstorage_volume" "bs-volume" {
  name = "bs-volume"
  size = 8
  volume_type = "ceph-hdd"
  availability_zone = "DP1"
}
```

# Create Blockstorage Snapshot
This example shows how to create a blockstorage snapshot from scratch.
First you have to set up basic configuration as shown in [Getting Started](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs/guides/getting_started)

Create blockstorage snapshot:

```hcl
resource "mcs_blockstorage_snapshot" "bs-snapshot" {
    name = "bs-snapshot"
    volume_id = example_volume_id
}
```
