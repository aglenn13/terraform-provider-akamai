provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

locals {
  entries = [for i in range(0, 1000) : "www.test.hostname.${i}.com.edgesuite.net"]
}

resource "akamai_property_hostname_bucket" "test" {
  property_id = "prp_111"
  contract_id = "ctr_222"
  group_id    = "grp_333"
  network     = "STAGING"
  hostnames = {
    for entry in concat(local.entries) :
    entry => {
      cert_provisioning_type = "DEFAULT"
      edge_hostname_id       = "ehn_444"
    }
  }
}
