# all basic info B with notifications C and grants
resource "akamai_iam_user" "test" {
  first_name     = "first name B"
  last_name      = "last name B"
  email          = "email-B@akamai.net"
  country        = "country B"
  phone          = "phone B"
  enable_tfa     = false
  send_otp_email = false

  contact_type       = "contact type B"
  job_title          = "job title B"
  time_zone          = "Timezone B"
  secondary_email    = "secondary-email-B@akamai.net"
  mobile_phone       = "mobile phone B"
  address            = "123 B Street"
  city               = "B-Town"
  state              = "state B"
  zip_code           = "zip B"
  preferred_language = "language B"
  session_timeout    = 2

  enable_notifications          = true
  subscribe_new_users           = true
  subscribe_password_expiration = true
  subscribe_product_issues      = ["issues product"]
  subscribe_product_upgrades    = ["upgrades product"]

  auth_grants_json = "[{\"groupId\":0,\"groupName\":\"B\",\"isBlocked\":false,\"roleDescription\":\"\",\"roleName\":\"\"}]"
}
