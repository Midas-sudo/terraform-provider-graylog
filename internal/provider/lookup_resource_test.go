// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLookupDataAdapterResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))[:8]
	name := "tf-adapter-" + suffix
	updatedName := "tf-adapter-upd-" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLookupDataAdapterResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_data_adapter.test", "name", name),
					resource.TestCheckResourceAttrSet("graylog_lookup_data_adapter.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_lookup_data_adapter.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLookupDataAdapterResourceConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_data_adapter.test", "name", updatedName),
				),
			},
		},
	})
}

func TestAccLookupCacheResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))[:8]
	name := "tf-cache-" + suffix
	updatedName := "tf-cache-upd-" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLookupCacheResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_cache.test", "name", name),
					resource.TestCheckResourceAttrSet("graylog_lookup_cache.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_lookup_cache.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLookupCacheResourceConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_cache.test", "name", updatedName),
				),
			},
		},
	})
}

func TestAccLookupTableResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLookupTableResourceConfig("tf-lktbl-"+suffix, "tf-lkcache-"+suffix, "tf-lkadapter-"+suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_table.test", "name", "tf-lktbl-"+suffix),
					resource.TestCheckResourceAttrSet("graylog_lookup_table.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_lookup_table.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLookupTableResourceConfig("tf-lktbl-upd-"+suffix, "tf-lkcache-"+suffix, "tf-lkadapter-"+suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_lookup_table.test", "name", "tf-lktbl-upd-"+suffix),
				),
			},
		},
	})
}

func TestAccLookupDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLookupDataSourcesConfig("tf-lktbl-ds-"+suffix, "tf-lkcache-ds-"+suffix, "tf-lkadapter-ds-"+suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_lookup_data_adapter.test", "id", "graylog_lookup_data_adapter.adapter", "id"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_data_adapters.test", "data_adapters.0.id"),
					resource.TestCheckResourceAttrPair("data.graylog_lookup_cache.test", "id", "graylog_lookup_cache.cache", "id"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_caches.test", "caches.0.id"),
					resource.TestCheckResourceAttrPair("data.graylog_lookup_table.test", "id", "graylog_lookup_table.test", "id"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_tables.test", "lookup_tables.0.id"),
				),
			},
		},
	})
}

func testAccLookupDataAdapterResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "graylog_lookup_data_adapter" "test" {
  title       = %[1]q
  name        = %[1]q
  description = "Terraform acceptance lookup adapter"
  custom_error_ttl_enabled = false
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
}
`, name)
}

func testAccLookupCacheResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "graylog_lookup_cache" "test" {
  title       = %[1]q
  name        = %[1]q
  description = "Terraform acceptance lookup cache"
  config = {
    type = "none"
  }
}
`, name)
}

func testAccLookupTableResourceConfig(tableName, cacheName, adapterName string) string {
	return fmt.Sprintf(`
resource "graylog_lookup_data_adapter" "adapter" {
  title       = %[3]q
  name        = %[3]q
  description = "Terraform acceptance lookup adapter for table"
  custom_error_ttl_enabled = false
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
}

resource "graylog_lookup_cache" "cache" {
  title       = %[2]q
  name        = %[2]q
  description = "Terraform acceptance lookup cache for table"
  config = {
    type = "none"
  }
}

resource "graylog_lookup_table" "test" {
  title                     = %[1]q
  name                      = %[1]q
  description               = "Terraform acceptance lookup table"
  cache_id                  = graylog_lookup_cache.cache.id
  data_adapter_id           = graylog_lookup_data_adapter.adapter.id
  default_single_value      = ""
  default_single_value_type = "NULL"
  default_multi_value       = "[]"
  default_multi_value_type  = "OBJECT"
}
`, tableName, cacheName, adapterName)
}

func testAccLookupDataSourcesConfig(tableName, cacheName, adapterName string) string {
	return fmt.Sprintf(`
%s

data "graylog_lookup_data_adapter" "test" {
  id = graylog_lookup_data_adapter.adapter.id
}

data "graylog_lookup_data_adapters" "test" {}

data "graylog_lookup_cache" "test" {
  id = graylog_lookup_cache.cache.id
}

data "graylog_lookup_caches" "test" {}

data "graylog_lookup_table" "test" {
  id = graylog_lookup_table.test.id
}

data "graylog_lookup_tables" "test" {}
`, testAccLookupTableResourceConfig(tableName, cacheName, adapterName))
}
