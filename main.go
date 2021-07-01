package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"gitlab.corp.mail.ru/infra/paas/terraform-provider-mcs/mcs"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: mcs.Provider})
}
