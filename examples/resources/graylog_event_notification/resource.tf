resource "graylog_event_notification" "example" {
  payload_json = jsonencode({
    title       = "Terraform Event Notification"
    description = "Managed by Terraform"
    config = {
      type = "http-notification-v1"
      url  = "https://example.org/graylog/events"
    }
  })
}
