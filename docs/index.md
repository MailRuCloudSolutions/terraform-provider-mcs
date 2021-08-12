---
layout: "mcs"
page_title: "Provider: MCS"
description: |-
  The MCS provider is used to interact with MCS services.
  The provider needs to be configured with the proper credentials before it can be used.
---

# MCS Provider

The MCS provider is used to interact with
[MCS services](https://mcs.mail.ru/). The provider needs
to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
# Configure the mcs provider

provider "mcs" {
    username   = "example@mail.ru"
    password   = "s3cr3t"
    project_id = "some_project_id"
}

# Create new kubernetes cluster
resource "mcs_kubernetes_cluster" "mycluster"{
  # ...
}
```

## Configuration Reference

The following arguments are supported:

* `username` - (Required) The username to login with.
  If omitted, the `USER_NAME` environment variable is used.

* `password` - (Required) The Password to login with. If omitted, the `PASSWORD` environment variable is used.

* `project_id` - (Required) The ID of Project to login with. 
  If omitted, the `PROJECT_ID` environment variable is used.

* `auth_url` - (Optional) URL for authentication in MCS. Default is https://infra.mail.ru/identity/v3/.

* `region` - (Optional) A region to use. Default is `RegionOne`. **New since v0.4.0**

