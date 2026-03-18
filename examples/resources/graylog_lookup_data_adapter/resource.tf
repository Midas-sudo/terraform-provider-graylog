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
