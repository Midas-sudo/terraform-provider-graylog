// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPluginTypesDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPluginTypesDataSourcesConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.graylog_input_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_input_types.test", "types.0.requested_configuration.0.name"),
					resource.TestCheckResourceAttrSet("data.graylog_output_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_output_types.test", "types.0.requested_configuration.0.name"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_adapter_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_adapter_types.test", "types.0.requested_configuration.0.name"),
					resource.TestCheckResourceAttrSet("data.graylog_lookup_cache_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_index_set_strategy_types.test", "rotation.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_index_set_strategy_types.test", "rotation.0.requested_configuration.0.name"),
					resource.TestCheckResourceAttrSet("data.graylog_index_set_strategy_types.test", "retention.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_event_notification_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_event_notification_types.test", "types.0.requested_configuration.0.name"),
					resource.TestCheckResourceAttrSet("data.graylog_event_entity_types.test", "processor_types.0"),
					resource.TestCheckResourceAttrSet("data.graylog_event_entity_types.test", "storage_handler_types.0"),
				),
			},
		},
	})
}

const testAccPluginTypesDataSourcesConfig = `
data "graylog_input_types" "test" {}
data "graylog_output_types" "test" {}
data "graylog_lookup_adapter_types" "test" {}
data "graylog_lookup_cache_types" "test" {}
data "graylog_index_set_strategy_types" "test" {}
data "graylog_event_notification_types" "test" {}
data "graylog_event_entity_types" "test" {}
`
