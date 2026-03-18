resource "graylog_lookup_cache" "example" {
  title       = "Terraform Lookup Cache"
  name        = "terraform-lookup-cache"
  description = "Managed by Terraform"

  config_json = jsonencode({
    type = "none"
  })
}
