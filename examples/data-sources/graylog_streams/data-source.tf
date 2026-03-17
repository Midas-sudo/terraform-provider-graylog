data "graylog_streams" "all" {}

output "stream_names" {
  value = [for s in data.graylog_streams.all.streams : s.title]
}
