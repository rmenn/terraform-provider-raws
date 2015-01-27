package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/rmenn/terraform-provider-raws/raws"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: raws.Provider,
	})
}
