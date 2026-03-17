data "graylog_pipelines" "all" {}

output "pipeline_names" {
  value = [for p in data.graylog_pipelines.all.pipelines : p.title]
}
