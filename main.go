package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"terraform-provider-multy/multy"
)

func main() {
	tfsdk.Serve(context.Background(), multy.New, tfsdk.ServeOpts{
		Name: "multy",
	})
}
