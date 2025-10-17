terraform {
  required_providers {
    active24 = {
      source  = "joystic/active24"
      version = "0.0.1"
    }
  }
}

provider "active24" {}

# resource "active24_dns_record" "a_example" {
#   domain   = "eternalstoic.com"   
#   service  = "10643001"           
#   name     = "tester6"             
#   type     = "A"
#   content  = "100.20.20.100"
#   ttl      = 300
# }


