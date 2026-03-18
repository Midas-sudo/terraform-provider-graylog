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
