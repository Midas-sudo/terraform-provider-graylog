data "graylog_index_template" "example" {
  index_set_id = "existing-index-set-id"
}

output "index_template_name" {
  value = data.graylog_index_template.example.name
}
