---
page_title: "active24_dns_record Resource"
subcategory: "DNS"
description: |-
  Manages a DNS record in an Active24 zone via REST API v2.
---

# active24_dns_record

Manages a DNS record in an Active24 zone. Supports all common record types including A, AAAA, CNAME, MX, TXT, SRV, and CAA.

## Example Usage

### A Record

```hcl
resource "active24_dns_record" "web" {
  domain  = "example.com"
  service = "12345678"
  name    = "www"
  type    = "A"
  content = "93.184.216.34"
  ttl     = 3600
}
```

### Multiple A Records (Round-Robin)

Multiple A records on the same name distribute traffic across several IPs.

```hcl
resource "active24_dns_record" "app_1" {
  domain  = "example.com"
  service = "12345678"
  name    = "app"
  type    = "A"
  content = "10.0.0.1"
  ttl     = 300
}

resource "active24_dns_record" "app_2" {
  domain  = "example.com"
  service = "12345678"
  name    = "app"
  type    = "A"
  content = "10.0.0.2"
  ttl     = 300
}
```

### AAAA Record (IPv6)

```hcl
resource "active24_dns_record" "ipv6" {
  domain  = "example.com"
  service = "12345678"
  name    = "www"
  type    = "AAAA"
  content = "2001:db8::1"
  ttl     = 3600
}
```

### CNAME Record

```hcl
resource "active24_dns_record" "blog" {
  domain  = "example.com"
  service = "12345678"
  name    = "blog"
  type    = "CNAME"
  content = "example.github.io"
  ttl     = 3600
}
```

### MX Record

```hcl
resource "active24_dns_record" "mail" {
  domain   = "example.com"
  service  = "12345678"
  name     = "@"
  type     = "MX"
  content  = "mail.example.com"
  priority = 10
  ttl      = 3600
}
```

### TXT Record

```hcl
resource "active24_dns_record" "spf" {
  domain  = "example.com"
  service = "12345678"
  name    = "@"
  type    = "TXT"
  content = "v=spf1 include:_spf.google.com ~all"
  ttl     = 3600
}
```

### CAA Record

CAA records restrict which Certificate Authorities may issue SSL/TLS certificates for the domain. Uses dedicated fields (`caa_flags`, `caa_tag`, `caa_value`) instead of `content`.

```hcl
# Allow Let's Encrypt to issue certificates for this domain
resource "active24_dns_record" "caa_issue" {
  domain    = "example.com"
  service   = "12345678"
  name      = "@"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "issue"
  caa_value = "letsencrypt.org"
  ttl       = 3600
}

# Allow Let's Encrypt to issue wildcard certificates
resource "active24_dns_record" "caa_issuewild" {
  domain    = "example.com"
  service   = "12345678"
  name      = "@"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "issuewild"
  caa_value = "letsencrypt.org"
  ttl       = 3600
}

# Report violations via email
resource "active24_dns_record" "caa_iodef" {
  domain    = "example.com"
  service   = "12345678"
  name      = "@"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "iodef"
  caa_value = "mailto:ssl@example.com"
  ttl       = 3600
}
```

### Using Azure Key Vault for Credentials

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

## Argument Reference

### Required

- `domain` - (String) Zone name (e.g. `example.com`).
- `name` - (String) Record name relative to the zone. Use `@` for the zone apex.
- `type` - (String) DNS record type. Supported: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `CAA`.

### Optional

- `service` - (String) Active24 service key. If omitted, the provider uses `domain`. Set this if your service ID in Active24 differs from the domain name.
- `content` - (String) Record value. **Required** for all record types except `CAA` (e.g. IP address for A, hostname for CNAME).
- `ttl` - (Number) Time-to-live in seconds. Defaults to `3600`.
- `priority` - (Number) Priority value for `MX` and `SRV` records.
- `caa_flags` - (Number) CAA record flags. Usually `0`. Only used when `type = "CAA"`.
- `caa_tag` - (String) CAA record tag. Only used when `type = "CAA"`. Valid values:
  - `issue` - authorize a CA to issue certificates for this domain
  - `issuewild` - authorize a CA to issue wildcard certificates
  - `iodef` - URL or email to report policy violations to
- `caa_value` - (String) CAA record value (e.g. `letsencrypt.org` or `mailto:ssl@example.com`). Only used when `type = "CAA"`.

## Attributes Reference

- `id` - (String) Unique record ID assigned by Active24.

## Import

Records can be imported using several formats. The provider auto-detects which format you use.

### Import by numeric ID

```bash
# Without service key
terraform import active24_dns_record.web "example.com:12345"

# With service key
terraform import active24_dns_record.web "example.com:12345678:12345"
```

### Import by name and type (recommended)

No need to know the numeric ID. The provider looks up the record via the API.

```bash
# Simple record (unique name+type combination)
terraform import active24_dns_record.web "example.com:12345678:www:A"

# CNAME record
terraform import active24_dns_record.blog "example.com:12345678:blog:CNAME"

# Full FQDN is also accepted - the domain suffix is stripped automatically
terraform import active24_dns_record.web "example.com:12345678:www.example.com:A"

# Apex record
terraform import active24_dns_record.mail "example.com:12345678:@:MX"
```

### Import with content disambiguation

When multiple records of the same type exist on the same name (e.g. multiple A records for round-robin, or multiple CAA records), add the content as the 5th part to select the correct record.

```bash
# Two A records on "app" - import each by its IP
terraform import active24_dns_record.app_1 "example.com:12345678:app:A:10.0.0.1"
terraform import active24_dns_record.app_2 "example.com:12345678:app:A:10.0.0.2"

# Multiple CAA records - import each by its value
terraform import active24_dns_record.caa_issue "example.com:12345678:@:CAA:letsencrypt.org"
terraform import active24_dns_record.caa_iodef "example.com:12345678:@:CAA:mailto:ssl@example.com"
```

-> **Tip:** If you attempt to import by name+type and multiple records match, the provider will list all matching records with their IDs and content so you can easily pick the right one.

## Import Format Summary

| Format | Example |
|--------|---------|
| `domain:id` | `example.com:12345` |
| `domain:service:id` | `example.com:12345678:12345` |
| `domain:name:type` | `example.com:www:A` |
| `domain:service:name:type` | `example.com:12345678:www:A` |
| `domain:service:name:type:content` | `example.com:12345678:app:A:10.0.0.1` |
