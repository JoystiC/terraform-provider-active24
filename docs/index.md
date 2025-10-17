---
page_title: "Active24 Provider"
subcategory: ""
description: |-
  Manage DNS records via Active24 REST API v2.
---

# Active24 Provider

The Active24 provider enables managing DNS records via the REST API v2.

## Authentication

Active24 v2 uses HMAC-signed Basic auth where:
- Username is your API key
- Password is a per-request HMAC-SHA1 signature of "METHOD path unix_timestamp"
- Header `X-Date` must contain the same timestamp in ISO8601 basic format

The provider signs requests for you. Set the following environment variables:

- `ACTIVE24_API_KEY`
- `ACTIVE24_API_SECRET`

Docs: [Active24 v2 Intro](https://rest.active24.cz/v2/docs/intro)

## Example Usage

```hcl
terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = ">= 0.0.1"
    }
  }
}

provider "active24" {}

resource "active24_dns_record" "example" {
  domain   = "example.com"
  service  = "10643001"   # or zone name if your account is configured that way
  name     = "www"
  type     = "A"
  content  = "192.0.2.10"
  ttl      = 300
}
```


