resource "graylog_content_pack" "example" {
  payload_json = jsonencode({
    id          = "00000000-0000-0000-0000-000000000001"
    v           = "1"
    rev         = 1
    name        = "Terraform Content Pack Example"
    summary     = "Example content pack created by Terraform"
    description = "Example content pack payload"
    vendor      = "Terraform"
    url         = "https://example.org"
    parameters  = []
    entities    = []
  })
}
