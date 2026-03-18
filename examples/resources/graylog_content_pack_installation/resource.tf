resource "graylog_content_pack_installation" "example" {
  content_pack_id = "90be5e03-cb16-c802-6462-a244b4a342f3"
  revision        = 1
  comment         = "Installed by Terraform"
  parameters_json = jsonencode({})
}
