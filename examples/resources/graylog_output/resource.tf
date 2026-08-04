resource "graylog_output" "example" {
  title = "Terraform Output"
  type  = "LoggingOutput"
  configuration = {
    prefix = "terraform-output:"
  }
}
