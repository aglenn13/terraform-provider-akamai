provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_iam_property_users" "test" {
  asset_id = "12345"
}
