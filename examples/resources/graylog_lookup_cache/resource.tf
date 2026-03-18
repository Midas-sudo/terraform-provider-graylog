resource "graylog_lookup_cache" "example" {
  payload_json = jsonencode({
    title       = "Terraform Lookup Cache"
    name        = "terraform-lookup-cache"
    description = "Managed by Terraform"
    config = {
      type = "none"
    }
  })
}
