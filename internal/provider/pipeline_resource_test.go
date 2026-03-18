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
