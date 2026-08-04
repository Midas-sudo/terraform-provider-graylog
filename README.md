# Terraform Provider for Graylog

A [Terraform](https://www.terraform.io) provider for managing [Graylog](https://graylog.org) configuration via its REST API.

Targets **Graylog 6.x** and designed for forward compatibility with **7.x**.

## Features (v0.8.0)

### Resources

| Resource | Description |
|---|---|
| `graylog_input` | Manage message inputs (Syslog, GELF, Beats, etc.) |
| `graylog_index_set` | Manage index sets, rotation and retention strategy settings |
| `graylog_stream` | Manage streams for message routing |
| `graylog_stream_rule` | Manage individual rules on a stream |
| `graylog_pipeline` | Manage processing pipelines |
| `graylog_pipeline_rule` | Manage pipeline rules |
| `graylog_pipeline_connection` | Connect pipelines to streams |
| `graylog_role` | Manage Graylog roles and permissions |
| `graylog_user` | Manage Graylog users (password write-only) |
| `graylog_entity_share` | Manage entity sharing capabilities for Graylog GRNs |
| `graylog_view` | Manage Graylog saved searches/views via raw JSON payload |
| `graylog_dashboard` | Manage Graylog dashboards via raw JSON payload |
| `graylog_event_definition` | Manage Graylog event definitions via raw JSON payload |
| `graylog_event_notification` | Manage Graylog event notifications via raw JSON payload |
| `graylog_event_definition_notification_binding` | Manage notification bindings attached to event definitions |
| `graylog_lookup_data_adapter` | Manage Graylog lookup data adapters via raw JSON payload |
| `graylog_lookup_cache` | Manage Graylog lookup caches via raw JSON payload |
| `graylog_lookup_table` | Manage Graylog lookup tables via raw JSON payload |
| `graylog_output` | Manage Graylog outputs via raw JSON payload |
| `graylog_extractor` | Manage Graylog input extractors via raw JSON payload |
| `graylog_grok_pattern` | Manage Graylog grok patterns via raw JSON payload |
| `graylog_content_pack` | Manage Graylog content pack revisions via raw JSON payload |
| `graylog_content_pack_installation` | Manage Graylog content pack installations |

### Data Sources

| Data Source | Description |
|---|---|
| `graylog_input` | Look up a single input by ID |
| `graylog_inputs` | List all inputs |
| `graylog_input_types` | List available input types |
| `graylog_index_set` | Look up a single index set by ID |
| `graylog_index_sets` | List all index sets |
| `graylog_index_template` | Read the generated index template for an index set |
| `graylog_stream` | Look up a single stream by ID |
| `graylog_streams` | List all streams |
| `graylog_pipeline` | Look up a single pipeline by ID |
| `graylog_pipelines` | List all pipelines |
| `graylog_pipeline_rule` | Look up a single pipeline rule by ID |
| `graylog_pipeline_rules` | List all pipeline rules |
| `graylog_role` | Look up a single role by name |
| `graylog_roles` | List all roles |
| `graylog_user` | Look up a single user by ID |
| `graylog_users` | List all users |
| `graylog_view` | Look up a single view by ID |
| `graylog_views` | List all saved views |
| `graylog_dashboard` | Look up a single dashboard by ID |
| `graylog_dashboards` | List all dashboards |
| `graylog_event_definition` | Look up a single event definition by ID |
| `graylog_event_definitions` | List all event definitions |
| `graylog_event_notification` | Look up a single event notification by ID |
| `graylog_event_notifications` | List all event notifications |
| `graylog_lookup_data_adapter` | Look up a single lookup data adapter by ID |
| `graylog_lookup_data_adapters` | List all lookup data adapters |
| `graylog_lookup_cache` | Look up a single lookup cache by ID |
| `graylog_lookup_caches` | List all lookup caches |
| `graylog_lookup_table` | Look up a single lookup table by ID |
| `graylog_lookup_tables` | List all lookup tables |
| `graylog_output` | Look up a single output by ID |
| `graylog_outputs` | List all outputs |
| `graylog_extractor` | Look up a single extractor by input and extractor IDs |
| `graylog_extractors` | List extractors for a specific input |
| `graylog_grok_patterns` | List all grok patterns |
| `graylog_content_pack` | Look up a single content pack by ID and revision |
| `graylog_content_packs` | List latest content pack revisions |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (to build the provider)
- A running Graylog instance with API access

## Provider Configuration

```hcl
terraform {
  required_providers {
    graylog = {
      source = "Midas-sudo/graylog"
    }
  }
}

provider "graylog" {
  endpoint = "https://graylog.example.com/api"
  username = "admin"
  password = "secret"

  # Or use token-based auth:
  # token = "your-api-token"

  # Optional:
  # insecure_skip_verify = true
  # timeout              = 60
}
```

All attributes can also be set via environment variables:

| Attribute | Environment Variable |
|---|---|
| `endpoint` | `GRAYLOG_ENDPOINT` |
| `username` | `GRAYLOG_USERNAME` |
| `password` | `GRAYLOG_PASSWORD` |
| `token` | `GRAYLOG_TOKEN` |

## Example Usage

```hcl
resource "graylog_input" "syslog" {
  title  = "Syslog UDP"
  type   = "org.graylog2.inputs.syslog.udp.SyslogUDPInput"
  global = true

  configuration = jsonencode({
    bind_address     = "0.0.0.0"
    port             = 1514
    recv_buffer_size = 262144
  })
}

resource "graylog_stream" "app_logs" {
  title                              = "Application Logs"
  index_set_id                       = "000000000000000000000001"
  matching_type                      = "AND"
  remove_matches_from_default_stream = true
}

resource "graylog_stream_rule" "by_source" {
  stream_id = graylog_stream.app_logs.id
  field     = "source"
  value     = "app-server-.*"
  type      = 2  # REGEX
}

resource "graylog_pipeline_rule" "extract" {
  source = <<-EOT
    rule "extract fields"
    when has_field("message")
    then set_field("processed", true);
    end
  EOT
}

resource "graylog_pipeline" "main" {
  source = <<-EOT
    pipeline "Main Pipeline"
    stage 0 match either
      rule "extract fields"
    end
  EOT
}

resource "graylog_pipeline_connection" "main_to_app" {
  stream_id    = graylog_stream.app_logs.id
  pipeline_ids = [graylog_pipeline.main.id]
}
```

## Building The Provider

```shell
go install
```

## Running Acceptance Tests

Acceptance tests run against a real Graylog instance and require `TF_ACC=1`.

### Local Graylog (Docker)

A local Graylog 7 + DataNode + MongoDB stack lives under `dev-workspace/` (gitignored personal workspace). Typical flow:

```shell
# Host prerequisite for DataNode
# echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf && sudo sysctl -p

cd dev-workspace
# compose sets GRAYLOG_SELFSIGNED_STARTUP=true to auto-complete DataNode preflight
docker compose up -d
# Wait until http://localhost:9000 accepts admin login, then return to the repo root.

cd ..
source scripts/local-acc-env.sh
make install
TF_ACC=1 make testacc
```

`scripts/local-acc-env.sh` exports `GRAYLOG_ENDPOINT`, `GRAYLOG_USERNAME`, `GRAYLOG_PASSWORD`, `TF_ACC`, and discovers `GRAYLOG_DEFAULT_INDEX_SET_ID` when the API is reachable.

Optional fixture overrides (auto-discovered when unset where supported):

| Variable | Purpose |
|---|---|
| `GRAYLOG_DEFAULT_INDEX_SET_ID` | Default index set for stream / share / index-set tests |
| `GRAYLOG_VIEW_SEARCH_ID` | Search ID for view/dashboard tests |
| `GRAYLOG_EVENT_STORAGE_STREAM_ID` | Storage stream for event definition tests |
| `GRAYLOG_CONTENT_PACK_ID` / `GRAYLOG_CONTENT_PACK_REVISION` | Existing pack for installation tests |

### Against any Graylog

```shell
export GRAYLOG_ENDPOINT="https://graylog.example.com/api"
export GRAYLOG_USERNAME="admin"
export GRAYLOG_PASSWORD="secret"
export TF_ACC=1

go test -v -count=1 ./internal/provider/ -timeout 120m
# or: make testacc
```

## Developing the Provider

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate` from the `tools/` directory.
