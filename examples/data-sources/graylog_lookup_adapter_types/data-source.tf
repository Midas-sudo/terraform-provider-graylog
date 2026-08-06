data "graylog_lookup_adapter_types" "available" {}

output "csvfile_default_config" {
  value = [
    for t in data.graylog_lookup_adapter_types.available.types : jsondecode(t.default_config)
    if t.type == "csvfile"
  ]
}
