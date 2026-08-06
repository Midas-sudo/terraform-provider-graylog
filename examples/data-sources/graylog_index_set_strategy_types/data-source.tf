data "graylog_index_set_strategy_types" "available" {}

output "rotation_strategies" {
  value = {
    for s in data.graylog_index_set_strategy_types.available.rotation :
    s.type => [for f in s.requested_configuration : f.name]
  }
}
