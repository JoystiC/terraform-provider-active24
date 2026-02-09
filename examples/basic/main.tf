terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "~> 1.1.0"
    }
  }
}

provider "active24" {}

# Import by ID:   terraform import active24_dns_record.a_example "finbricks.com:12905048:311059968"
# Import by name:  terraform import active24_dns_record.a_example "finbricks.com:12905048:devtest:A"
resource "active24_dns_record" "a_example" {
  domain  = "finbricks.com"
  service = "12905048"
  name    = "devtest"
  type    = "A"
  content = "1.2.3.5"
  ttl     = 3600
}

# Import by ID:   terraform import active24_dns_record.cname_example "finbricks.com:12905048:123456789"
# Import by name:  terraform import active24_dns_record.cname_example "finbricks.com:12905048:devtest.dev:CNAME"
resource "active24_dns_record" "cname_example" {
  domain  = "finbricks.com"
  service = "12905048"
  name    = "devtest.dev"
  type    = "CNAME"
  content = "grafana.dev.finbricks.com"
  ttl     = 3600
}

# Import by ID:   terraform import active24_dns_record.caa_example "finbricks.com:12905048:311059866"
# Import by name:  terraform import active24_dns_record.caa_example "finbricks.com:12905048:devtest.dev:CAA"
resource "active24_dns_record" "caa_example" {
  domain    = "finbricks.com"
  service   = "12905048"
  name      = "devtest.dev"
  type      = "CAA"
  caa_flags = 0
  caa_tag   = "issue"
  caa_value = "letsencrypt.org"
  ttl       = 3600
}
