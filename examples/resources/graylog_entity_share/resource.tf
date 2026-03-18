resource "graylog_entity_share" "example" {
  entity_grn = "grn::::stream:000000000000000000000001"

  grantee_capabilities = {
    "grn::::builtin-team:everyone" = "view"
  }
}
