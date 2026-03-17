data "graylog_pipeline_rules" "all" {}

output "rule_names" {
  value = [for r in data.graylog_pipeline_rules.all.rules : r.title]
}
