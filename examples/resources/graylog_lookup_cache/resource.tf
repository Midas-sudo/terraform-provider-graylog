resource "graylog_lookup_cache" "none" {
  title       = "Terraform Null Cache"
  name        = "terraform-lookup-cache-none"
  description = "Managed by Terraform"

  config = {
    type = "none"
  }
}

resource "graylog_lookup_cache" "guava" {
  title       = "Terraform Guava Cache"
  name        = "terraform-lookup-cache-guava"
  description = "Managed by Terraform"

  config = {
    type                     = "guava_cache"
    max_size                 = 1000
    expire_after_access      = 60
    expire_after_access_unit = "SECONDS"
    expire_after_write       = 0
    ignore_null              = false
  }
}
