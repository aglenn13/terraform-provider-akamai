provider "akamai" {
  edgerc        = "../../common/testutils/edgerc"
  cache_enabled = false
}

data "akamai_appsec_rapid_rules" "test" {
  security_policy_id = "2222_333333"
}
