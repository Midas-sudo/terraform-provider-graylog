resource "graylog_pipeline_rule" "extract_fields" {
  source = <<-EOT
    rule "extract fields"
    when
      has_field("message")
    then
      set_field("processed", true);
    end
  EOT

  description = "Extract and set fields on messages"
}
