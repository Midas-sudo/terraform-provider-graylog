resource "graylog_index_set" "example" {
  title                    = "Application Index Set"
  description              = "Managed by Terraform"
  index_prefix             = "app"
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "TimeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"

  rotation_strategy = {
    type                   = "TimeBasedRotationStrategyConfig"
    rotation_period        = "P1D"
    rotate_empty_index_set = false
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }

  sync_template = true
}

resource "graylog_index_set" "size_based" {
  title                    = "Size-Based Index Set"
  description              = "Rotate when index reaches ~1 GiB"
  index_prefix             = "sized"
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "SizeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"

  rotation_strategy = {
    type     = "SizeBasedRotationStrategyConfig"
    max_size = 1073741824
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }

  sync_template = true
}

resource "graylog_index_set" "size_optimizing" {
  title                    = "Optimized Index Set"
  description              = "Time-based size optimizing with data tiering"
  index_prefix             = "opt"
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = false
  rotation_strategy_class  = "TimeBasedSizeOptimizingStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"

  rotation_strategy = {
    type = "TimeBasedSizeOptimizingStrategyConfig"
  }

  retention_strategy = {
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
