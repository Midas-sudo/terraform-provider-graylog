data "graylog_inputs" "all" {}

output "input_names" {
  value = [for i in data.graylog_inputs.all.inputs : i.title]
}
