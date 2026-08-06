## 0.12.0 (Aug 6, 2026)

FEATURES:

- Type-discovery data sources expose Graylog plugin configuration fields (`requested_configuration`, defaults, optional/required) so Dynamic resource attrs are discoverable at plan time:
  - `graylog_input_types` now returns full field metadata from `/system/inputs/types/all`
  - `graylog_output_types`
  - `graylog_lookup_adapter_types` / `graylog_lookup_cache_types`
  - `graylog_index_set_strategy_types` (rotation + retention)
  - `graylog_event_notification_types` (modern built-ins + legacy alarm callbacks)
  - `graylog_event_entity_types` (processor / field provider / storage / aggregation catalogs)
- Registry docs for Dynamic resources include per-type configuration tables and point at the matching types data source.

## 0.11.1 (Aug 4, 2026)

BUG FIXES:

- `graylog_index_set`: pass through the full Graylog `rotation_strategy` and `retention_strategy` objects (Dynamic attributes) instead of only `type` / `max_number_of_indices`. This restores fields such as `rotation_period`, `max_docs_per_index`, `max_size`, and `rotate_empty_index_set` on create/update.

BREAKING CHANGES:

- `rotation_strategy` and `retention_strategy` changed from nested blocks to required Dynamic attributes (HCL objects). Existing state is migrated via schema version `0 → 1`.

  Before:

  ```hcl
  rotation_strategy {
    type = "TimeBasedRotationStrategyConfig"
  }
  ```

  After:

  ```hcl
  rotation_strategy = {
    type                   = "TimeBasedRotationStrategyConfig"
    rotation_period        = "P1D"
    rotate_empty_index_set = false
  }
  ```

## 0.11.0 (Aug 4, 2026)

BREAKING CHANGES:

- Plugin configuration attributes now use native HCL objects (`schema.DynamicAttribute`) instead of JSON strings. Existing state is migrated automatically via schema version upgraders.

  Before:

  ```hcl
  configuration = jsonencode({
    bind_address = "0.0.0.0"
    port         = 1514
  })
  ```

  After:

  ```hcl
  configuration = {
    bind_address = "0.0.0.0"
    port         = 1514
  }
  ```

- Attribute renames (update Terraform configs; state upgrades rename keys):
  - `graylog_output.configuration_json` → `configuration`
  - `graylog_lookup_data_adapter.config_json` → `config`
  - `graylog_lookup_cache.config_json` → `config`
  - `graylog_event_notification.config_json` → `config`
  - `graylog_extractor.extractor_config_json` / `converters_json` → `extractor_config` / `converters`
  - `graylog_event_definition.config_json` / `field_spec_json` / `notifications_json` / `storage_json` → `config` / `field_spec` / `notifications` / `storage`
  - `graylog_event_definition.notification_settings_json` → typed nested attribute `notification_settings` (`grace_period_ms`, `backlog_size`)

- `graylog_input.configuration` type changes from JSON string to Dynamic object (name unchanged).

NOTES:

- List data sources keep nested config fields as JSON strings where the Plugin Framework forbids Dynamic inside collections; singular data sources mirror resource Dynamic types.
- View/dashboard `*_json` and content-pack `entities_json` / `parameters_json` remain JSON strings.

## 0.10.1 (Mar 20, 2026)

ENHANCEMENTS:

- Add import examples for all resources under `examples/resources/*/import.sh`.
- Regenerate docs so every resource page includes an `Import` section with correct ID format examples.

## 0.10.0 (Mar 20, 2026)

FEATURES:

- Add short-name alias support for index set strategy classes and config types.
- Add short-name alias support for input and output types.
- Keep backward compatibility with fully-qualified Graylog class names.

ENHANCEMENTS:

- Update provider docs and examples to prefer short type names.
- Extend acceptance and unit coverage for alias expansion/collapse behavior.
