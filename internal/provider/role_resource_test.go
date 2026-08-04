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

func TestAccRoleResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	roleName := "tf-role-" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResourceConfig(roleName, "Terraform acceptance role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_role.test", "name", roleName),
					resource.TestCheckResourceAttr("graylog_role.test", "description", "Terraform acceptance role"),
					resource.TestCheckResourceAttrSet("graylog_role.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleResourceConfig(roleName, "Terraform acceptance role updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_role.test", "description", "Terraform acceptance role updated"),
				),
			},
		},
	})
}

func TestAccRoleDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	roleName := "tf-role-ds-" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourcesConfig(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.graylog_role.test", "name", roleName),
					resource.TestCheckResourceAttrSet("data.graylog_role.test", "description"),
					resource.TestCheckResourceAttrSet("data.graylog_roles.test", "roles.0.name"),
				),
			},
		},
	})
}

func testAccRoleResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "graylog_role" "test" {
  name        = %[1]q
  description = %[2]q
  permissions = [
    "streams:read",
    "streams:edit",
  ]
}
`, name, description)
}

func testAccRoleDataSourcesConfig(name string) string {
	return fmt.Sprintf(`
%s

data "graylog_role" "test" {
  name = graylog_role.test.name
}

data "graylog_roles" "test" {}
`, testAccRoleResourceConfig(name, "Terraform acceptance role data source"))
}
