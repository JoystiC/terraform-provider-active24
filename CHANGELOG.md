## v1.1.0

### New Features
- **Import by name and type**: Import existing DNS records without knowing their numeric ID. Supported formats:
  - `<domain>:<name>:<type>` (e.g. `finbricks.com:www:A`)
  - `<domain>:<service>:<name>:<type>` (e.g. `finbricks.com:12905048:devtest.dev:CAA`)
  - Full FQDN is also accepted (e.g. `finbricks.com:12905048:devtest.finbricks.com:A`)
- **CAA record support** with dedicated fields (`caa_value`, `caa_flags`, `caa_tag`).

### Bug Fixes
- Fixed FQDN drift: the provider now strips the domain suffix from record names returned by the API.
- Fixed CAA content synthesis: `content` is auto-populated from `caa_value` for Active24 API compatibility.
- Fixed CAA field consistency: `Read` now correctly populates all CAA fields after create, update, and import.
- Fixed CNAME/A 500 errors: CAA-specific fields are no longer sent for non-CAA record types.
- Validation: `content` is now required for non-CAA records.

## v0.0.17

- Initial release of terraform-provider-active24 (community provider)
- Provider configuration with HMAC-signed Basic auth (api_key + api_secret)
- Resource `active24_dns_record` supporting A/AAAA/CNAME/TXT/MX/SRV (generic fields)
- Post-create/update readback to capture ID and TTL; supports import


