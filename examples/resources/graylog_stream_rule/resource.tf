resource "graylog_stream_rule" "source_match" {
  stream_id   = graylog_stream.application_logs.id
  field       = "source"
  value       = "app-server-.*"
  type        = 2 # REGEX
  inverted    = false
  description = "Match messages from application servers"
}
