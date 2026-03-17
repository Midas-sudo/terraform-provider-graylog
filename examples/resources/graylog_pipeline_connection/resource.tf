resource "graylog_pipeline_connection" "example" {
  stream_id    = graylog_stream.application_logs.id
  pipeline_ids = [graylog_pipeline.example.id]
}
