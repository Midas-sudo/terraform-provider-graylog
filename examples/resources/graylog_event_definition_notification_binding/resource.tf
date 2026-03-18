resource "graylog_event_notification" "example" {
  title       = "Terraform Event Notification"
  description = "Managed by Terraform"
  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/graylog/events"
  })
}

resource "graylog_event_definition" "example" {
  title       = "Terraform Event Definition"
  description = "Managed by Terraform"
  priority    = 1
  alert       = false
  config_json = jsonencode({
    type = "system-notifications-v1"
  })
  field_spec_json = jsonencode({})
  key_spec        = []
  notification_settings_json = jsonencode({
    grace_period_ms = 0
    backlog_size    = 0
  })
  notifications_json = jsonencode([])
  storage_json = jsonencode([
    {
      type    = "persist-to-streams-v1"
      streams = ["000000000000000000000003"]
    }
  ])
  state = "ENABLED"
}

resource "graylog_event_definition_notification_binding" "example" {
  event_definition_id = graylog_event_definition.example.id
  notification_ids    = [graylog_event_notification.example.id]
}
