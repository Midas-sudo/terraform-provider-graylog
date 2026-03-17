data "graylog_input_types" "available" {}

output "available_input_types" {
  value = [for t in data.graylog_input_types.available.types : "${t.name} (${t.type})"]
}
