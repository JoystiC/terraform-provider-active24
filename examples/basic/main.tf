terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "~> 1.0.5"
    }
  }
}

provider "active24" {}


# terraform import active24_dns_record.caa_example "finbricks.com:12905048:311059866"
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

# terraform import active24_dns_record.a_example "finbricks.com:12905048:311059968"
resource "active24_dns_record" "a_example" {
  domain  = "finbricks.com"
  name    = "devtest"
  type    = "A"
  content = "1.2.3.5"
  ttl     = 3600
  service   = "12905048"
}

# terraform import active24_dns_record.cname_example "finbricks.com:12905048:123456789"
resource "active24_dns_record" "cname_example" {
  domain  = "finbricks.com"
  name    = "devtest.dev"
  type    = "CNAME"
  content = "grafana.dev.finbricks.com"
  ttl     = 3600
  service   = "12905048"
}
