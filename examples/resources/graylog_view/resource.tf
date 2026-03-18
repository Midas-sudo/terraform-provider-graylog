resource "graylog_view" "example" {
  payload_json = jsonencode({
    type        = "SEARCH"
    title       = "Terraform View"
    summary     = "Search view managed by Terraform"
    description = "A minimal Graylog view payload"
    search_id   = "existing-search-id"
    properties  = []
    requires    = {}
    state       = {}
  })
}
