resource "graylog_event_notification" "example" {
  title       = "Terraform Event Notification"
  description = "Managed by Terraform"

  config = {
    type = "http-notification-v1"
    url  = "https://example.org/graylog/events"
  }
}
