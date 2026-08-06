data "graylog_output_types" "available" {}

output "output_type_fields" {
  value = {
    for t in data.graylog_output_types.available.types :
    t.type => [for f in t.requested_configuration : f.name]
  }
}
