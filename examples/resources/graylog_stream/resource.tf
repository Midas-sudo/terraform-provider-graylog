resource "graylog_stream" "application_logs" {
  title                              = "Application Logs"
  description                        = "All application log messages"
  index_set_id                       = "000000000000000000000001"
  matching_type                      = "AND"
  remove_matches_from_default_stream = true
}
