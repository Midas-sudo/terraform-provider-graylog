// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccInputResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInputResourceConfig("Test Syslog UDP"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_input.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Syslog UDP"),
					),
					statecheck.ExpectKnownValue(
						"graylog_input.test",
						tfjsonpath.New("global"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"graylog_input.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("SyslogUDPInput"),
					),
				},
			},
			{
				ResourceName:            "graylog_input.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
			{
				Config: testAccInputResourceConfig("Test Syslog UDP Updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_input.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Syslog UDP Updated"),
					),
				},
			},
		},
	})
}

func TestAccInputDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInputDataSourcesConfig("TF Acc Input DS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_input.test", "id", "graylog_input.test", "id"),
					resource.TestCheckResourceAttr("data.graylog_input.test", "title", "TF Acc Input DS"),
					resource.TestCheckResourceAttrSet("data.graylog_inputs.test", "inputs.0.id"),
					resource.TestCheckResourceAttrSet("data.graylog_input_types.test", "types.0.type"),
					resource.TestCheckResourceAttrSet("data.graylog_input_types.test", "types.0.requested_configuration.0.name"),
				),
			},
		},
	})
}

func testAccInputResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_input" "test" {
  title  = %[1]q
  type   = "SyslogUDPInput"
  global = true

  configuration = {
    bind_address     = "0.0.0.0"
    port             = 15140
    recv_buffer_size = 262144
  }
}
`, title)
}

func testAccInputDataSourcesConfig(title string) string {
	return fmt.Sprintf(`
%s

data "graylog_input" "test" {
  id = graylog_input.test.id
}

data "graylog_inputs" "test" {}

data "graylog_input_types" "test" {}
`, testAccInputResourceConfig(title))
}
