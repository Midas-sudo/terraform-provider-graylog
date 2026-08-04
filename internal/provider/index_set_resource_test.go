// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIndexSetResource(t *testing.T) {
	testAccRequireDefaultIndexSetID(t)

	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	prefix := "tfidx" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexSetResourceConfig(prefix, "Terraform index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "index_prefix", prefix),
					resource.TestCheckResourceAttr("graylog_index_set.test", "title", "Terraform index set"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.type", "MessageCountRotationStrategyConfig"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.max_docs_per_index", "20000000"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "retention_strategy.max_number_of_indices", "20"),
					resource.TestCheckResourceAttrSet("graylog_index_set.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_index_set.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"set_as_default",
					"sync_template",
					"is_default",
					// Dynamic strategy objects may include API-normalized number encodings.
					"rotation_strategy",
					"retention_strategy",
				},
			},
			{
				Config: testAccIndexSetResourceConfig(prefix, "Terraform index set updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "title", "Terraform index set updated"),
				),
			},
		},
	})
}

func TestAccIndexSetResourceTimeBasedRotation(t *testing.T) {
	testAccRequireDefaultIndexSetID(t)

	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	prefix := "tftbr" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexSetResourceTimeBasedConfig(prefix, "Terraform time-based index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy_class", "TimeBasedRotationStrategy"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.type", "TimeBasedRotationStrategyConfig"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.rotation_period", "P1D"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.rotate_empty_index_set", "false"),
				),
			},
			{
				Config: testAccIndexSetResourceTimeBasedConfigUpdated(prefix, "Terraform time-based index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.rotation_period", "P7D"),
				),
			},
		},
	})
}

func TestAccIndexSetResourceSizeBasedRotation(t *testing.T) {
	testAccRequireDefaultIndexSetID(t)

	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	prefix := "tfsbr" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexSetResourceSizeBasedConfig(prefix, "Terraform size-based index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy_class", "SizeBasedRotationStrategy"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.type", "SizeBasedRotationStrategyConfig"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.max_size", "1073741824"),
				),
			},
			{
				Config: testAccIndexSetResourceSizeBasedConfigUpdated(prefix, "Terraform size-based index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.max_size", "2147483648"),
				),
			},
		},
	})
}

func TestAccIndexSetDataSources(t *testing.T) {
	defaultIndexSetID := testAccRequireDefaultIndexSetID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexSetDataSourcesConfig(defaultIndexSetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.graylog_index_set.test", "id", defaultIndexSetID),
					resource.TestCheckResourceAttrSet("data.graylog_index_set.test", "title"),
					resource.TestCheckResourceAttrSet("data.graylog_index_sets.test", "index_sets.0.id"),
					resource.TestCheckResourceAttrSet("data.graylog_index_template.test", "name"),
				),
			},
		},
	})
}

func TestAccIndexSetResourceDataTiering(t *testing.T) {
	testAccRequireDefaultIndexSetID(t)

	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	prefix := "tfdt" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexSetResourceDataTieringConfig(prefix, "Terraform data tiering index set"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_index_set.test", "index_prefix", prefix),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy_class", "TimeBasedSizeOptimizingStrategy"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "rotation_strategy.type", "TimeBasedSizeOptimizingStrategyConfig"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "data_tiering.type", "hot_only"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "data_tiering.index_lifetime_min", "P30D"),
					resource.TestCheckResourceAttr("graylog_index_set.test", "data_tiering.index_lifetime_max", "P40D"),
				),
			},
			{
				ResourceName:      "graylog_index_set.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"set_as_default",
					"sync_template",
					"is_default",
					"rotation_strategy",
					"retention_strategy",
				},
			},
		},
	})
}

func testAccRequireDefaultIndexSetID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if id == "" {
		t.Skip("GRAYLOG_DEFAULT_INDEX_SET_ID must be set for index set acceptance tests")
	}
	return id
}

func testAccIndexSetResourceConfig(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  rotation_strategy_class  = "MessageCountRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

  rotation_strategy = {
    type               = "MessageCountRotationStrategyConfig"
    max_docs_per_index = 20000000
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }
}
`, title, indexPrefix)
}

func testAccIndexSetResourceTimeBasedConfig(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test (time-based rotation)"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "TimeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

  rotation_strategy = {
    type                   = "TimeBasedRotationStrategyConfig"
    rotation_period        = "P1D"
    rotate_empty_index_set = false
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 30
  }
}
`, title, indexPrefix)
}

func testAccIndexSetResourceTimeBasedConfigUpdated(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test (time-based rotation)"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "TimeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

  rotation_strategy = {
    type                   = "TimeBasedRotationStrategyConfig"
    rotation_period        = "P7D"
    rotate_empty_index_set = false
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 30
  }
}
`, title, indexPrefix)
}

func testAccIndexSetResourceSizeBasedConfig(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test (size-based rotation)"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "SizeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

  rotation_strategy = {
    type     = "SizeBasedRotationStrategyConfig"
    max_size = 1073741824
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }
}
`, title, indexPrefix)
}

func testAccIndexSetResourceSizeBasedConfigUpdated(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test (size-based rotation)"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = true
  rotation_strategy_class  = "SizeBasedRotationStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

  rotation_strategy = {
    type     = "SizeBasedRotationStrategyConfig"
    max_size = 2147483648
  }

  retention_strategy = {
    type                  = "DeletionRetentionStrategyConfig"
    max_number_of_indices = 20
  }
}
`, title, indexPrefix)
}

func testAccIndexSetDataSourcesConfig(indexSetID string) string {
	return fmt.Sprintf(`
data "graylog_index_set" "test" {
  id = %[1]q
}

data "graylog_index_sets" "test" {}

data "graylog_index_template" "test" {
  index_set_id = %[1]q
}
`, indexSetID)
}

func testAccIndexSetResourceDataTieringConfig(indexPrefix, title string) string {
	return fmt.Sprintf(`
resource "graylog_index_set" "test" {
  title                    = %[1]q
  description              = "Managed by acceptance test (data tiering)"
  index_prefix             = %[2]q
  shards                   = 1
  replicas                 = 0
  writable                 = true
  index_analyzer           = "standard"
  use_legacy_rotation      = false
  rotation_strategy_class  = "TimeBasedSizeOptimizingStrategy"
  retention_strategy_class = "DeletionRetentionStrategy"
  set_as_default           = false
  sync_template            = true

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
}
`, title, indexPrefix)
}
