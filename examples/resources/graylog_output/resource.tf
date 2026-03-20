resource "graylog_output" "example" {
  title = "Terraform Output"
  type  = "LoggingOutput"
  configuration_json = jsonencode({
    prefix = "terraform-output:"
  })
}
