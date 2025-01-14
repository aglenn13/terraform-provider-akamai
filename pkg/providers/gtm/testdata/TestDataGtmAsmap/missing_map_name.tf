provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_gtm_asmap" "my_gtm_asmap" {
  domain = "gtm_terra_testdomain.akadns.net"
}
