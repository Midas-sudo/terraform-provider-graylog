// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStreamResource(t *testing.T) {
	indexSetID := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if indexSetID == "" {
		indexSetID = "000000000000000000000001"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamResourceConfig("Test Stream", indexSetID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_stream.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Stream"),
					),
					statecheck.ExpectKnownValue(
						"graylog_stream.test",
						tfjsonpath.New("matching_type"),
						knownvalue.StringExact("AND"),
					),
				},
			},
			{
				ResourceName:      "graylog_stream.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamResourceConfig("Test Stream Updated", indexSetID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_stream.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Stream Updated"),
					),
				},
			},
		},
	})
}

func TestAccStreamRuleResource(t *testing.T) {
	indexSetID := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if indexSetID == "" {
		indexSetID = "000000000000000000000001"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamRuleResourceConfig(indexSetID, "source", "app-server", "Match app servers"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("field"),
						knownvalue.StringExact("source"),
					),
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("value"),
						knownvalue.StringExact("app-server"),
					),
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("type"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("inverted"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "graylog_stream_rule.test",
				ImportState:       true,
				ImportStateIdFunc: testAccStreamRuleImportIDFunc("graylog_stream_rule.test"),
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamRuleResourceConfig(indexSetID, "source", "app-server-updated", "Match updated app servers"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("value"),
						knownvalue.StringExact("app-server-updated"),
					),
					statecheck.ExpectKnownValue(
						"graylog_stream_rule.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Match updated app servers"),
					),
				},
			},
		},
	})
}

func TestAccStreamDataSources(t *testing.T) {
	indexSetID := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if indexSetID == "" {
		indexSetID = "000000000000000000000001"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamDataSourcesConfig("TF Acc Stream DS", indexSetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_stream.test", "id", "graylog_stream.test", "id"),
					resource.TestCheckResourceAttr("data.graylog_stream.test", "title", "TF Acc Stream DS"),
					resource.TestCheckResourceAttrSet("data.graylog_streams.test", "streams.0.id"),
				),
			},
		},
	})
}

func testAccStreamResourceConfig(title, indexSetID string) string {
	return fmt.Sprintf(`
resource "graylog_stream" "test" {
  title                              = %[1]q
  description                        = "Terraform acceptance test stream"
  index_set_id                       = %[2]q
  matching_type                      = "AND"
  remove_matches_from_default_stream = false
}
`, title, indexSetID)
}

func testAccStreamDataSourcesConfig(title, indexSetID string) string {
	return fmt.Sprintf(`
%s

data "graylog_stream" "test" {
  id = graylog_stream.test.id
}

data "graylog_streams" "test" {}
`, testAccStreamResourceConfig(title, indexSetID))
}

func testAccStreamRuleResourceConfig(indexSetID, field, value, description string) string {
	return fmt.Sprintf(`
resource "graylog_stream" "test" {
  title                              = "Test Stream for Rule"
  description                        = "Terraform acceptance test stream for stream rule"
  index_set_id                       = %[1]q
  matching_type                      = "AND"
  remove_matches_from_default_stream = false
}

resource "graylog_stream_rule" "test" {
  stream_id   = graylog_stream.test.id
  field       = %[2]q
  value       = %[3]q
  type        = 1
  inverted    = false
  description = %[4]q
}
`, indexSetID, field, value, description)
}

func testAccStreamRuleImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		streamID := rs.Primary.Attributes["stream_id"]
		ruleID := rs.Primary.ID
		if streamID == "" || ruleID == "" {
			return "", fmt.Errorf("missing stream_id or rule id in state")
		}
		return streamID + "/" + ruleID, nil
	}
}
