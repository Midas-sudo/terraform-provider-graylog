data "graylog_inputs" "all" {}

resource "graylog_extractor" "example" {
  input_id = data.graylog_inputs.all.inputs[0].id
  payload_json = jsonencode({
    title           = "Terraform Extractor"
    source_field    = "message"
    target_field    = "message_copy"
    extractor_type  = "copy_input"
    cursor_strategy = "copy"
    condition_type  = "none"
    condition_value = ""
    extractor_config = {}
    converters       = []
    order            = 0
  })
}
