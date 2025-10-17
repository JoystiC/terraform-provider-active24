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
```

## Argument Reference

- `domain` (Required) Zone name.
- `service` (Optional) Active24 v2 service key. If omitted, the provider uses `domain`.
- `name` (Required) Record name. Use `@` for apex.
- `type` (Required) Record type (A, AAAA, CNAME, TXT, MX, SRV, ...).
- `content` (Required) Record value.
- `ttl` (Optional) Time-to-live in seconds. Defaults to 3600.
- `priority` (Optional) Priority (MX/SRV).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Record ID assigned by Active24.

## Import

Import using `<domain>:<id>`:

```bash
terraform import active24_dns_record.a_example example.com:12345
```


