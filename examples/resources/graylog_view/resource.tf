resource "graylog_view" "example" {
  title       = "Terraform View"
  summary     = "Search view managed by Terraform"
  description = "A minimal Graylog view payload"
  search_id   = "existing-search-id"

  properties_json = jsonencode([])
  requires_json   = jsonencode({})
  state_json      = jsonencode({})
}
