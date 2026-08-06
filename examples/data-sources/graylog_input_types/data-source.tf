data "graylog_input_types" "available" {}

output "syslog_udp_fields" {
  value = [
    for t in data.graylog_input_types.available.types : {
      type   = t.type
      fields = [for f in t.requested_configuration : "${f.name} (optional=${f.is_optional})"]
    }
    if t.type == "SyslogUDPInput"
  ]
}
