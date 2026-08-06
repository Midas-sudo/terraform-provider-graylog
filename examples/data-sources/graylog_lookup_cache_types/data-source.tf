data "graylog_lookup_cache_types" "available" {}

output "cache_types" {
  value = [for t in data.graylog_lookup_cache_types.available.types : t.type]
}
