package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	"terraform-provider-multy/multy"
)

func main() {
	err := providerserver.Serve(context.Background(), multy.New, providerserver.ServeOpts{
		Address: "hashicorp.com/dev/multy",
	})
	if err != nil {
		log.Printf("unable to start provider, %s", err)
	}
}
