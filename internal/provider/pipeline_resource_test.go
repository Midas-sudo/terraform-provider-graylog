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
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPipelineResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineResourceConfig("Test Pipeline"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_pipeline.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Pipeline"),
					),
				},
			},
			{
				ResourceName:      "graylog_pipeline.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineResourceConfig("Test Pipeline Updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_pipeline.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Pipeline Updated"),
					),
				},
			},
		},
	})
}

func TestAccPipelineRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineRuleResourceConfig("test rule"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_pipeline_rule.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("test rule"),
					),
				},
			},
			{
				ResourceName:      "graylog_pipeline_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineRuleResourceConfig("test rule updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"graylog_pipeline_rule.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("test rule updated"),
					),
				},
			},
		},
	})
}

func TestAccPipelineDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDataSourcesConfig("TF Acc Pipeline DS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_pipeline.test", "id", "graylog_pipeline.test", "id"),
					resource.TestCheckResourceAttrSet("data.graylog_pipelines.test", "pipelines.0.id"),
				),
			},
		},
	})
}

func TestAccPipelineRuleDataSources(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineRuleDataSourcesConfig("tf acc pipeline rule ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_pipeline_rule.test", "id", "graylog_pipeline_rule.test", "id"),
					resource.TestCheckResourceAttrSet("data.graylog_pipeline_rules.test", "rules.0.id"),
				),
			},
		},
	})
}

func TestAccPipelineConnectionResource(t *testing.T) {
	indexSetID := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if indexSetID == "" {
		indexSetID = "000000000000000000000001"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConnectionResourceConfig(indexSetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("graylog_pipeline_connection.test", "id"),
					resource.TestCheckResourceAttrSet("graylog_pipeline_connection.test", "stream_id"),
					resource.TestCheckResourceAttr("graylog_pipeline_connection.test", "pipeline_ids.#", "1"),
				),
			},
			{
				ResourceName:      "graylog_pipeline_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineConnectionResourceConfigUpdated(indexSetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_pipeline_connection.test", "pipeline_ids.#", "2"),
				),
			},
		},
	})
}

func testAccPipelineResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_pipeline" "test" {
  source = <<-EOT
    pipeline "%[1]s"
    stage 0 match either
    end
  EOT

  description = "Terraform acceptance test pipeline"
}
`, title)
}

func testAccPipelineConnectionResourceConfig(indexSetID string) string {
	return fmt.Sprintf(`
resource "graylog_stream" "test" {
  title                              = "Test Stream for Pipeline Connection"
  description                        = "Terraform acceptance test stream for pipeline connection"
  index_set_id                       = %[1]q
  matching_type                      = "AND"
  remove_matches_from_default_stream = false
}

resource "graylog_pipeline" "test" {
  source = <<-EOT
    pipeline "Test Pipeline for Connection"
    stage 0 match either
    end
  EOT

  description = "Terraform acceptance test pipeline for connection"
}

resource "graylog_pipeline_connection" "test" {
  stream_id    = graylog_stream.test.id
  pipeline_ids = [graylog_pipeline.test.id]
}
`, indexSetID)
}

func testAccPipelineConnectionResourceConfigUpdated(indexSetID string) string {
	return fmt.Sprintf(`
resource "graylog_stream" "test" {
  title                              = "Test Stream for Pipeline Connection"
  description                        = "Terraform acceptance test stream for pipeline connection"
  index_set_id                       = %[1]q
  matching_type                      = "AND"
  remove_matches_from_default_stream = false
}

resource "graylog_pipeline" "test" {
  source = <<-EOT
    pipeline "Test Pipeline for Connection"
    stage 0 match either
    end
  EOT

  description = "Terraform acceptance test pipeline for connection"
}

resource "graylog_pipeline" "second" {
  source = <<-EOT
    pipeline "Second Pipeline for Connection"
    stage 0 match either
    end
  EOT

  description = "Second Terraform acceptance test pipeline for connection"
}

resource "graylog_pipeline_connection" "test" {
  stream_id    = graylog_stream.test.id
  pipeline_ids = [graylog_pipeline.test.id, graylog_pipeline.second.id]
}
`, indexSetID)
}

func testAccPipelineRuleResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_pipeline_rule" "test" {
  source = <<-EOT
    rule "%[1]s"
    when
      has_field("message")
    then
      set_field("processed_by_terraform", true);
    end
  EOT

  description = "Terraform acceptance test rule"
}
`, title)
}

func testAccPipelineDataSourcesConfig(title string) string {
	return fmt.Sprintf(`
%s

data "graylog_pipeline" "test" {
  id = graylog_pipeline.test.id
}

data "graylog_pipelines" "test" {}
`, testAccPipelineResourceConfig(title))
}

func testAccPipelineRuleDataSourcesConfig(title string) string {
	return fmt.Sprintf(`
%s

data "graylog_pipeline_rule" "test" {
  id = graylog_pipeline_rule.test.id
}

data "graylog_pipeline_rules" "test" {}
`, testAccPipelineRuleResourceConfig(title))
}
