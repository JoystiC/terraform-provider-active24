---
page_title: "Active24 Provider"
subcategory: ""
description: |-
  Manage DNS records via the Active24 REST API v2.
---

# Active24 Provider

The Active24 provider enables managing DNS records hosted at [Active24](https://www.active24.cz/) via the REST API v2.

It supports creating, updating, deleting, and importing DNS records of all common types including A, AAAA, CNAME, MX, TXT, SRV, and CAA.

## Authentication

Active24 API v2 uses HMAC-signed Basic authentication. The provider handles request signing automatically - you only need to provide your API key and secret.

### Option 1: Environment Variables (Recommended)

```bash
export ACTIVE24_API_KEY="your-api-key"
export ACTIVE24_API_SECRET="your-api-secret"
```

```hcl
provider "active24" {}
```

### Option 2: Explicit Configuration

```hcl
provider "active24" {
  api_key    = "your-api-key"
  api_secret = "your-api-secret"
}
```

### Option 3: Azure Key Vault Integration

```hcl
data "azurerm_key_vault_secret" "api_key" {
  name         = "active24-api-key"
  key_vault_id = data.azurerm_key_vault.kv.id
}

data "azurerm_key_vault_secret" "api_secret" {
  name         = "active24-api-secret"
  key_vault_id = data.azurerm_key_vault.kv.id
}

provider "active24" {
  api_key    = data.azurerm_key_vault_secret.api_key.value
  api_secret = data.azurerm_key_vault_secret.api_secret.value
}
```

You can obtain your API credentials from the [Active24 administration panel](https://customer.active24.com/).

API documentation: [Active24 REST v2](https://rest.active24.cz/v2/docs/intro)

## Example Usage

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

# A record
resource "active24_dns_record" "web" {
  domain  = "example.com"
  service = "12345678"
  name    = "www"
  type    = "A"
  content = "93.184.216.34"
  ttl     = 3600
}

# CNAME record
resource "active24_dns_record" "blog" {
  domain  = "example.com"
  service = "12345678"
  name    = "blog"
  type    = "CNAME"
  content = "example.github.io"
  ttl     = 3600
}

# CAA record - restrict SSL certificate issuance
resource "active24_dns_record" "caa" {
  domain    = "example.com"
  service   = "12345678"
  name      = "@"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "issue"
  caa_value = "letsencrypt.org"
  ttl       = 3600
}
```

## Schema

### Optional

- `api_key` - (String) Active24 API key. Can also be set via `ACTIVE24_API_KEY` environment variable.
- `api_secret` - (String, Sensitive) Active24 API secret used to sign requests. Can also be set via `ACTIVE24_API_SECRET` environment variable.
- `base_url` - (String) Base URL for the Active24 API. Defaults to `https://rest.active24.cz/v2`.
