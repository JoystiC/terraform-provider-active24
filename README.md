# Terraform Provider for Active24 DNS

Community Terraform provider for managing DNS records at [Active24](https://www.active24.cz/) via their REST API v2.

[![Registry](https://img.shields.io/badge/terraform-registry-blue)](https://registry.terraform.io/providers/JoystiC/active24/latest)

## Features

- Manage DNS records: **A**, **AAAA**, **CNAME**, **MX**, **TXT**, **SRV**, **CAA**
- Full **CAA support** with dedicated fields (`caa_flags`, `caa_tag`, `caa_value`)
- **Smart import** - import existing records by name and type, no numeric ID needed
- Content-based disambiguation for multiple records on the same name (round-robin A, multiple CAA)
- HMAC-signed authentication handled automatically

## Quick Start

```hcl
terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "~> 1.1"
    }
  }
}

provider "active24" {}

resource "active24_dns_record" "web" {
  domain  = "example.com"
  service = "12345678"
  name    = "www"
  type    = "A"
  content = "93.184.216.34"
  ttl     = 3600
}
```

## Authentication

Set your API credentials via environment variables:

```bash
export ACTIVE24_API_KEY="your-api-key"
export ACTIVE24_API_SECRET="your-api-secret"
```

Or pass them explicitly in the provider block, or load them from Azure Key Vault. See [provider documentation](https://registry.terraform.io/providers/JoystiC/active24/latest/docs) for details.

## Import

Import existing DNS records without knowing their numeric ID:

```bash
# By name and type
terraform import active24_dns_record.web "example.com:12345678:www:A"

# With content disambiguation (multiple records on same name)
terraform import active24_dns_record.app_1 "example.com:12345678:app:A:10.0.0.1"

# CAA records
terraform import active24_dns_record.caa "example.com:12345678:@:CAA:letsencrypt.org"

# By numeric ID (also supported)
terraform import active24_dns_record.web "example.com:12345678:98765"
```

## Local Development

```bash
# Build and install locally
make build
make install VERSION=1.1.0

# Run tests
cd examples/basic
export ACTIVE24_API_KEY="..."
export ACTIVE24_API_SECRET="..."
terraform plan
```

For rapid iteration, use [dev_overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) in `~/.terraformrc` to skip `terraform init`:

```hcl
provider_installation {
  dev_overrides {
    "joystic/active24" = "/path/to/bin"
  }
  direct {}
}
```

## API Reference

- [Active24 REST v2 Docs](https://rest.active24.cz/v2/docs)
- [OpenAPI Spec](https://rest.active24.cz/v2/docs/openapi.json)

## License

See [LICENSE](LICENSE).
