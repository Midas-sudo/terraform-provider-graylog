# Terraform Provider for Graylog

A [Terraform](https://www.terraform.io) provider for managing [Graylog](https://graylog.org) configuration via its REST API.

Targets **Graylog 6.x** and designed for forward compatibility with **7.x**.

## Features (v0.1)

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

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (to build the provider)
- A running Graylog instance with API access

## Provider Configuration

```hcl
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

Acceptance tests run against a real Graylog instance. Set the required environment variables and run:

```shell
export GRAYLOG_ENDPOINT="https://graylog.example.com/api"
export GRAYLOG_USERNAME="admin"
export GRAYLOG_PASSWORD="secret"

go test -v -count=1 ./internal/provider/ -timeout 120m
```

## Developing the Provider

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate` from the `tools/` directory.
