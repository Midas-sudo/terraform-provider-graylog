resource "graylog_index_set" "example" {
  title                    = "Application Index Set"
  description              = "Managed by Terraform"
  index_prefix             = "app"
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = false
  rotation_strategy_class  = "TimeBasedSizeOptimizingStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"

  rotation_strategy {
    type = "TimeBasedSizeOptimizingStrategyConfig"
  }

  retention_strategy {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }

  data_tiering {
    type               = "hot_only"
    index_lifetime_min = "P30D"
    index_lifetime_max = "P40D"
  }

  sync_template = true
}
