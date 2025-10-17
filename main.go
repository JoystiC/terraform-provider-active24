package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/JoystiC/terraform-provider-active24/internal/provider"
)

// version is set by goreleaser or at build time via -ldflags
var version = "dev"

func main() {
	ctx := context.Background()

	providerserver.Serve(ctx, provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/joystic/active24",
	})
}
