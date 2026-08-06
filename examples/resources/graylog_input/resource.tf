resource "graylog_input" "syslog_udp" {
  title  = "Syslog UDP"
  type   = "SyslogUDPInput"
  global = true

  configuration = {
    bind_address           = "0.0.0.0"
    port                   = 1514
    recv_buffer_size       = 262144
    allow_override_date    = true
    store_full_message     = false
    force_rdns             = false
    expand_structured_data = false
  }

  static_fields = {
    from_terraform = "true"
  }
}

resource "graylog_input" "gelf_udp" {
  title  = "GELF UDP"
  type   = "GELFUDPInput"
  global = true

  configuration = {
    bind_address          = "0.0.0.0"
    port                  = 12201
    recv_buffer_size      = 262144
    decompress_size_limit = 8388608
    number_worker_threads = 2
  }
}
