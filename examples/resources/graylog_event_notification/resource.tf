resource "graylog_event_notification" "example" {
  title       = "Terraform Event Notification"
  description = "Managed by Terraform"

  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/graylog/events"
  })
}
