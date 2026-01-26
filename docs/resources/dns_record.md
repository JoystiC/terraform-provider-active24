---
page_title: "active24_dns_record Resource"
subcategory: "DNS"
description: |-
  Manages an Active24 DNS record.
---

# active24_dns_record

Manages a DNS record in an Active24 zone.

## Example Usage

```hcl
resource "active24_dns_record" "a_example" {
  domain   = "example.com"
  service  = "10643001" # or the zone name depending on your account
  name     = "@"
  type     = "A"
  content  = "203.0.113.10"
  ttl      = 300
}

resource "active24_dns_record" "caa_example" {
  domain    = "example.com"
  name      = "@"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "issue"
  caa_value = "letsencrypt.org"
}
```

## Argument Reference

- `domain` (Required) Zone name (e.g., `example.com`).
- `service` (Optional) Active24 v2 service key. If omitted, the provider uses `domain`. Use this if your service ID in Active24 differs from the domain name.
- `name` (Required) Record name. Use `@` for the zone apex.
- `type` (Required) Record type. Supported types include `A`, `AAAA`, `CNAME`, `TXT`, `MX`, `SRV`, `CAA`.
- `content` (Optional) Record value. **Required** for all record types except `CAA`.
- `ttl` (Optional) Time-to-live in seconds. Defaults to `3600`.
- `priority` (Optional) Priority value. Used for `MX` and `SRV` records.
- `caa_flags` (Optional) Criticality flags for `CAA` records (usually `0`).
- `caa_tag` (Optional) Tag for `CAA` records. Common values: `issue`, `issuewild`, `iodef`.
- `caa_value` (Optional) The value for the `CAA` record (e.g., the CA domain like `letsencrypt.org`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Record ID assigned by Active24.

## Import

Import using `<domain>:<id>`:

```bash
terraform import active24_dns_record.a_example example.com:12345
```


