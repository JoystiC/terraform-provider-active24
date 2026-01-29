## v1.0.5

- Fixed FQDN drift in record names: the provider now automatically strips the domain suffix when reading from the API.

## v1.0.1

- Added support for **CAA records** with specialized fields (`caa_value`, `caa_flags`, `caa_tag`).
- Enhanced validation for DNS records: `content` is now strictly required for non-CAA records.
- Updated documentation with CAA examples and detailed argument descriptions.
- Improved resource logic to handle server-side TTL normalization and record lookup consistency.

## v0.0.17

- Initial release of terraform-provider-active24 (community provider)
- Provider configuration with HMAC-signed Basic auth (api_key + api_secret)
- Resource `active24_dns_record` supporting A/AAAA/CNAME/TXT/MX/SRV (generic fields)
- Post-create/update readback to capture ID and TTL; supports import


