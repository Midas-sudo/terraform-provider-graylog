resource "graylog_output" "example" {
  title = "Terraform Output"
  type  = "org.graylog2.outputs.LoggingOutput"
  configuration_json = jsonencode({
    prefix = "terraform-output:"
  })
}
