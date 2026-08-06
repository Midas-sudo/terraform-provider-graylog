data "graylog_event_entity_types" "available" {}

output "event_processors" {
  value = data.graylog_event_entity_types.available.processor_types
}
