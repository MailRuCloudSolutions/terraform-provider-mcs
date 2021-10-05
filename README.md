Terraform MCS Provider
============================

* Documentation [https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs](https://registry.terraform.io/providers/MailRuCloudSolutions/mcs/latest/docs)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.0.x
-	[Go](https://golang.org/doc/install) 1.17 (to build the provider plugin)

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
To start improve it grab the repository, build it and install into local registry repository.
Builds for MacOS, Windows and Linux are available.
The example is for MacOS.
```sh
$ mkdir -p $GOPATH/src/github.com/MailRuCloudSolutions
$ cd $GOPATH/src/github.com/MailRuCloudSolutions
$ git clone git@github.com:MailRuCloudSolutions/terraform-provider-mcs.git
$ cd $GOPATH/src/github.com/MailRuCloudSolutions/terraform-provider-mcs
$ make build_darwin
$ mdkir -p ~/.terraform.d/plugins/hub.mcs.mail.ru/repository/mcs/0.5.3/darwin_amd64/
$ cp terraform-provider-mcs_darwin ~/.terraform.d/plugins/hub.mcs.mail.ru/repository/mcs/0.5.3/darwin_amd64/terraform-provider-mcs_v0.5.3

$ cat <<EOF > main.tf 
terraform {
  required_providers {
    mcs = {
      source  = "hub.mcs.mail.ru/repository/mcs"
      version = "0.5.3"
    }
  }
}
EOF
$ terraform init
```

Publishing provider
-------------------
Provider publishes via action [release](https://github.com/MailRuCloudSolutions/terraform-provider-mcs/blob/master/.github/workflows/release.yml).
To call the action create new tag.
```sh
$ git tag v0.5.3
$ git push origin v0.5.3
```

Thank You!