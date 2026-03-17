resource "graylog_pipeline" "example" {
  source = <<-EOT
    pipeline "Example Pipeline"
    stage 0 match either
      rule "extract fields"
    end
  EOT

  description = "An example processing pipeline"
}
