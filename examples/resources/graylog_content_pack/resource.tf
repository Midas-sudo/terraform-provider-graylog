resource "graylog_content_pack" "example" {
  content_pack_id = "00000000-0000-0000-0000-000000000001"
  v               = "1"
  revision        = 1
  name            = "Terraform Content Pack Example"
  summary         = "Example content pack created by Terraform"
  description     = "Example content pack payload"
  vendor          = "Terraform"
  url             = "https://example.org"
  parameters_json = jsonencode([])
  entities_json   = jsonencode([])
}
