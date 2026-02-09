terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "~> 1.3"
    }
  }
}

provider "active24" {
  # Credentials are loaded from environment variables:
  #   ACTIVE24_API_KEY
  #   ACTIVE24_API_SECRET
  #
  # Or set them explicitly:
  #   api_key    = "your-api-key"
  #   api_secret = "your-api-secret"
}

# --- A record ---
# Points a subdomain to an IP address.
resource "active24_dns_record" "web" {
  domain  = "example.com"
  service = "12345678"
  name    = "www"
  type    = "A"
  content = "93.184.216.34"
  ttl     = 3600
}

# --- Multiple A records on the same name (round-robin) ---
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

# --- CNAME record ---
# Alias pointing to another domain name.
resource "active24_dns_record" "blog" {
  domain  = "example.com"
  service = "12345678"
  name    = "blog"
  type    = "CNAME"
  content = "example.github.io"
  ttl     = 3600
}

# --- MX record ---
# Mail exchange record with priority.
resource "active24_dns_record" "mail" {
  domain   = "example.com"
  service  = "12345678"
  name     = "@"
  type     = "MX"
  content  = "mail.example.com"
  priority = 10
  ttl      = 3600
}

# --- TXT record ---
# Commonly used for SPF, DKIM, domain verification, etc.
resource "active24_dns_record" "spf" {
  domain  = "example.com"
  service = "12345678"
  name    = "@"
  type    = "TXT"
  content = "v=spf1 include:_spf.google.com ~all"
  ttl     = 3600
}

# --- CAA record ---
# Restricts which Certificate Authorities can issue SSL certificates.
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

# --- Multiple CAA records on the same name ---
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
