package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/MailRuCloudSolutions/terraform-provider-mcs/mcs"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mcs.Provider})
}
