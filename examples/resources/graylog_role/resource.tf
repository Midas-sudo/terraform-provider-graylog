resource "graylog_role" "example" {
  name        = "terraform_role_example"
  description = "Role managed by Terraform"
  permissions = [
    "streams:read",
    "streams:edit",
  ]
}
