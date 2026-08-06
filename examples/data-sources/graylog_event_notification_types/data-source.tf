data "graylog_event_notification_types" "available" {}

output "http_notification_fields" {
  value = [
    for t in data.graylog_event_notification_types.available.types : [
      for f in t.requested_configuration : "${f.name} (optional=${f.is_optional})"
    ]
    if t.type == "http-notification-v1"
  ]
}
