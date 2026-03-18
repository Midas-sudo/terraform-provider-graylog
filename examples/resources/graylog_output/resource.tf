resource "graylog_output" "example" {
  payload_json = jsonencode({
    title = "Terraform Output"
    type  = "org.graylog2.outputs.LoggingOutput"
    configuration = {
      prefix = "terraform-output:"
    }
  })
}
