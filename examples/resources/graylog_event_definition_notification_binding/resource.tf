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

resource "graylog_event_definition" "example" {
  payload_json = jsonencode({
    title       = "Terraform Event Definition"
    description = "Managed by Terraform"
    priority    = 1
    alert       = false
    config = {
      type = "system-notifications-v1"
    }
    field_spec = {}
    key_spec   = []
    notification_settings = {
      grace_period_ms = 0
      backlog_size    = 0
    }
    notifications = []
    storage = [
      {
        type    = "persist-to-streams-v1"
        streams = ["000000000000000000000003"]
      }
    ]
    scheduler = null
    state     = "ENABLED"
  })
}

resource "graylog_event_definition_notification_binding" "example" {
  event_definition_id = graylog_event_definition.example.id
  notification_ids    = [graylog_event_notification.example.id]
}
