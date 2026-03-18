resource "graylog_dashboard" "example" {
  payload_json = jsonencode({
    type        = "DASHBOARD"
    title       = "Terraform Dashboard"
    summary     = "Dashboard managed by Terraform"
    description = "A minimal Graylog dashboard payload"
    search_id   = "existing-search-id"
    properties  = []
    requires    = {}
    state       = {}
  })
}
