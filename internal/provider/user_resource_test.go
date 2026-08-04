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

func TestAccUserResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	username := "tf-user-" + suffix[:8]
	email := username + "@example.local"
	emailUpdated := username + "-updated@example.local"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_user.test", "username", username),
					resource.TestCheckResourceAttr("graylog_user.test", "email", email),
					resource.TestCheckResourceAttrSet("graylog_user.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccUserResourceConfig(username, emailUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_user.test", "email", emailUpdated),
				),
			},
		},
	})
}

func TestAccUserDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	username := "tf-user-ds-" + suffix[:8]
	email := username + "@example.local"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourcesConfig(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_user.test", "id", "graylog_user.test", "id"),
					resource.TestCheckResourceAttr("data.graylog_user.test", "username", username),
					resource.TestCheckResourceAttrSet("data.graylog_users.test", "users.0.id"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(username, email string) string {
	return fmt.Sprintf(`
resource "graylog_user" "test" {
  username = %[1]q
  password = "ChangeMe123!"
  email    = %[2]q

  first_name = "Terraform"
  last_name  = "User"

  roles = ["Reader"]
}
`, username, email)
}

func testAccUserDataSourcesConfig(username, email string) string {
	return fmt.Sprintf(`
%s

data "graylog_user" "test" {
  id = graylog_user.test.id
}

data "graylog_users" "test" {}
`, testAccUserResourceConfig(username, email))
}
