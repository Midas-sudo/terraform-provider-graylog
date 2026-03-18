resource "graylog_grok_pattern" "example" {
  payload_json = jsonencode({
    name    = "TFEXAMPLEPATTERN"
    pattern = "foo(?<bar>.*)"
  })
}
