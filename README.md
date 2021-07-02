Terraform MCS Provider
============================

* Documentation [https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.0.x
-	[Go](https://golang.org/doc/install) 1.16 (to build the provider plugin)

Using The Provider
----------------------
To use the provider, prepare configuration files based on examples from [here](https://github.com/MailRuCloudSolutions/terraform-provider-mcs/tree/master/examples)

```sh
$ cd $GOPATH/src/github.com/MailRuCloudSolutions/terraform-provider-mcs/examples/create-mcs-cluster
$ vim provider.tf
$ terraform init
$ terraform plan
```

Provider development
---------------------
```sh
$ mkdir -p $GOPATH/src/github.com/MailRuCloudSolutions
$ cd $GOPATH/src/github.com/MailRuCloudSolutions
$ git clone git@github.com:MailRuCloudSolutions/terraform-provider-mcs.git
$ cd $GOPATH/src/github.com/MailRuCloudSolutions/terraform-provider-mcs
$ make build
...
$ $GOPATH/bin/terraform-provider-mcs
```

Thank You