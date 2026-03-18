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

func TestAccEntityShareResource(t *testing.T) {
	indexSetID := os.Getenv("GRAYLOG_DEFAULT_INDEX_SET_ID")
	if indexSetID == "" {
		t.Skip("GRAYLOG_DEFAULT_INDEX_SET_ID must be set for entity share acceptance tests")
	}

	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	username := "tf-share-user-" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEntityShareResourceConfig(indexSetID, username, "view"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("graylog_entity_share.test", "id"),
					resource.TestCheckResourceAttr("graylog_entity_share.test", "grantee_capabilities.%", "1"),
				),
			},
			{
				ResourceName:      "graylog_entity_share.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEntityShareResourceConfig(indexSetID, username, "manage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_entity_share.test", "grantee_capabilities.%", "1"),
				),
			},
		},
	})
}

func testAccEntityShareResourceConfig(indexSetID, username, capability string) string {
	return fmt.Sprintf(`
resource "graylog_user" "share_user" {
  username = %[1]q
  password = "ChangeMe123!"
  email    = "%[1]s@example.local"
  first_name = "Terraform"
  last_name  = "Share"
  roles    = ["Reader"]
}

resource "graylog_stream" "share_stream" {
  title                              = "TF Share Stream %[1]s"
  description                        = "Terraform acceptance stream for entity sharing"
  index_set_id                       = %[2]q
  matching_type                      = "AND"
  remove_matches_from_default_stream = false
}

resource "graylog_entity_share" "test" {
  entity_grn = "grn::::stream:${graylog_stream.share_stream.id}"
  grantee_capabilities = {
    "grn::::user:${graylog_user.share_user.id}" = %[3]q
  }
}
`, username, indexSetID, capability)
}
