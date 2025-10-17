## v0.0.1

- Initial release of terraform-provider-active24 (community provider)
- Provider configuration with HMAC-signed Basic auth (api_key + api_secret)
- Resource `active24_dns_record` supporting A/AAAA/CNAME/TXT/MX/SRV (generic fields)
- Post-create/update readback to capture ID and TTL; supports import


