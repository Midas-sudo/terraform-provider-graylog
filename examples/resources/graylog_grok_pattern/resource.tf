resource "graylog_grok_pattern" "example" {
  name    = "TFEXAMPLEPATTERN"
  pattern = "foo(?<bar>.*)"
}
