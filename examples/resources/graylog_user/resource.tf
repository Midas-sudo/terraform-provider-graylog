resource "graylog_user" "example" {
  username = "terraform_user_example"
  password = "replace-with-secure-password"
  email    = "terraform@example.local"

  first_name = "Terraform"
  last_name  = "User"

  roles = [
    "Reader",
  ]
}
