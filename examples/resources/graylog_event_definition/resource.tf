resource "graylog_event_definition" "example" {
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
  state = "ENABLED"
}
