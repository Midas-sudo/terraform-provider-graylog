resource "graylog_lookup_data_adapter" "example" {
  payload_json = jsonencode({
    title       = "Terraform Lookup Adapter"
    name        = "terraform-lookup-adapter"
    description = "Managed by Terraform"
    config = {
      type                    = "csvfile"
      path                    = "/tmp/lookup-table.csv"
      separator               = ","
      quotechar               = "\""
      key_column              = "key"
      value_column            = "value"
      check_interval          = 60
      case_insensitive_lookup = false
      multi_value_lookup      = false
      cidr_lookup             = false
    }
  })
}

resource "graylog_lookup_cache" "example" {
  payload_json = jsonencode({
    title       = "Terraform Lookup Cache"
    name        = "terraform-lookup-cache"
    description = "Managed by Terraform"
    config = {
      type = "none"
    }
  })
}

resource "graylog_lookup_table" "example" {
  payload_json = jsonencode({
    title                     = "Terraform Lookup Table"
    name                      = "terraform-lookup-table"
    description               = "Managed by Terraform"
    cache_id                  = graylog_lookup_cache.example.id
    data_adapter_id           = graylog_lookup_data_adapter.example.id
    default_single_value      = ""
    default_single_value_type = "NULL"
    default_multi_value       = "[]"
    default_multi_value_type  = "OBJECT"
  })
}
