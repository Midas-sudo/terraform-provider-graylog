data "graylog_index_sets" "all" {}

output "index_set_titles" {
  value = [for s in data.graylog_index_sets.all.index_sets : s.title]
}
