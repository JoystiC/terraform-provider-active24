## Terraform Provider for Active24 DNS

This is a community Terraform provider for managing DNS records at Active24 via their REST API.

API reference: [Active24 REST v2 Docs](https://rest.active24.cz/v2/docs) | [OpenAPI JSON](https://rest.active24.cz/v2/docs/openapi.json)

### Installation

Local development install:

```bash
make tidy
make install VERSION=0.0.1
```

This installs the provider binary to `~/.terraform.d/plugins/registry.terraform.io/petrskyva/active24/0.0.1/${OS}_${ARCH}/terraform-provider-active24`.

### Authentication

Active24 v2 uses HMAC-signed Basic auth: username is API key, password is signature per request, and header `X-Date` must match the signature timestamp. The provider handles signing; set:

```bash
export ACTIVE24_API_KEY=...    # from Admin
export ACTIVE24_API_SECRET=... # from Admin
```

### Usage

See `examples/basic/main.tf`:

```hcl
terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "0.0.1"
    }
  }
}

provider "active24" {}

resource "active24_dns_record" "a_example" {
  domain  = "example.com"
  # optional: if your service key differs from the domain, set it explicitly
  # service = "example.com:host"  # example format if applicable
  name    = "@"
  type    = "A"
  content = "1.2.3.4"
  ttl     = 3600
}
```

If you use Terragrunt, point modules to this provider the same way as Terraform, and pass the token via an environment variable like `ACTIVE24_API_TOKEN` [[memory:6829071]].

### Notes

- Endpoints used were inferred from the DNS section and may need alignment with exact paths/fields in the API. Adjust `internal/provider/client.go` accordingly.
- Import format: `<domain>:<id>` (e.g., `example.com:12345`).


