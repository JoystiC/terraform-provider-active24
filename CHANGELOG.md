## v1.3.0

### New Features
- **CAA record support** with dedicated fields (`caa_value`, `caa_flags`, `caa_tag`).
- **Import by name and type**: Import existing DNS records without knowing their numeric ID. Supported formats:
  - `<domain>:<name>:<type>` (e.g. `example.com:www:A`)
  - `<domain>:<service>:<name>:<type>` (e.g. `example.com:12345678:www:CAA`)
  - `<domain>:<service>:<name>:<type>:<content>` for disambiguation (e.g. `example.com:12345678:app:A:10.0.0.1`)
  - Full FQDN is also accepted (e.g. `example.com:12345678:www.example.com:A`)
- **Content-based disambiguation**: When multiple records of the same type exist on the same name, import by content to select the correct one. If ambiguous, the provider lists all matching records with their IDs.
- **Validation**: `content` is now required for non-CAA records.
- **Comprehensive documentation** with examples for all record types (A, AAAA, CNAME, MX, TXT, CAA).

### Bug Fixes
- Fixed FQDN drift: the provider now strips the domain suffix from record names returned by the API.
- Fixed CAA content synthesis: `content` is auto-populated from `caa_value` for Active24 API compatibility.
- Fixed CAA field consistency: `Read` now correctly populates all CAA fields after create, update, and import.
- Fixed CNAME/A 500 errors: CAA-specific fields are no longer sent for non-CAA record types.
- Fixed TTL drift: the provider preserves TTL from configuration when the API returns 0.
- Fixed record read-back after create: content-based matching prevents picking the wrong record when duplicates exist.

## v0.0.17

- Initial release of terraform-provider-active24 (community provider)
- Provider configuration with HMAC-signed Basic auth (api_key + api_secret)
- Resource `active24_dns_record` supporting A/AAAA/CNAME/TXT/MX/SRV (generic fields)
- Post-create/update readback to capture ID and TTL; supports import
