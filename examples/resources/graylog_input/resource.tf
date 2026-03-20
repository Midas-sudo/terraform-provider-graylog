resource "graylog_input" "syslog_udp" {
  title  = "Syslog UDP"
  type   = "SyslogUDPInput"
  global = true

  configuration = jsonencode({
    bind_address           = "0.0.0.0"
    port                   = 1514
    recv_buffer_size       = 262144
    allow_override_date    = true
    store_full_message     = false
    force_rdns             = false
    expand_structured_data = false
  })

  static_fields = {
    from_terraform = "true"
  }
}
