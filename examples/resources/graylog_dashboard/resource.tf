resource "graylog_dashboard" "example" {
  title       = "Terraform Dashboard"
  summary     = "Dashboard managed by Terraform"
  description = "A minimal Graylog dashboard payload"
  search_id   = "existing-search-id"

  properties_json = jsonencode([])
  requires_json   = jsonencode({})
  state_json      = jsonencode({})
}
