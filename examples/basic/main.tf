terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "~> 1.0.4"
    }
  }
}

provider "active24" {}

# Příklad pro A záznam (IP adresa)
# resource "active24_dns_record" "a_example" {
#   domain  = "mojedomena.cz"
#   name    = "www"
#   type    = "A"
#   content = "1.2.3.4"
#   ttl     = 3600
# }

# # Příklad pro CNAME záznam (alias)
# resource "active24_dns_record" "cname_example" {
#   domain  = "mojedomena.cz"
#   name    = "blog"
#   type    = "CNAME"
#   content = "ghs.googlehost.com."
#   ttl     = 3600
# }

# # Příklad pro CAA záznam (SSL certifikace)
# resource "active24_dns_record" "caa_example" {
#   domain    = "mojedomena.cz"
#   name      = "@" # apex doména
#   type      = "CAA"
#   caa_flags = 0
#   caa_tag   = "issue"
#   caa_value = "letsencrypt.org"
#   ttl       = 3600
# }

# resource "active24_dns_record" "caa_example" {
#   domain    = "finbricks.com"
#   service   = "12905048"
#   name      = "devtest"
#   type      = "CAA"
#   caa_flags = 0
#   caa_tag   = "issue"
#   caa_value = "letsencrypt.org"
#   # content zde již není potřeba
#   ttl       = 3600
# }

# resource "active24_dns_record" "a_example" {
#   domain  = "finbricks.com"
#   name    = "devtest"
#   type    = "A"
#   content = "1.2.3.4"
#   ttl     = 3600
#   service   = "12905048"
# }

# # Příklad pro CNAME záznam (alias)
# resource "active24_dns_record" "cname_example" {
#   domain  = "finbricks.com"
#   name    = "devtest"
#   type    = "CNAME"
#   content = "api.dev.finbricks.com"
#   ttl     = 3600
#   service   = "12905048"
# }
