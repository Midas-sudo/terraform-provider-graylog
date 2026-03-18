// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccViewResource(t *testing.T) {
	searchID := testAccResolveSearchID(t)
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "TF View " + suffix[:8]
	updatedTitle := title + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewResourceConfig(title, searchID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_view.test", "type", "SEARCH"),
					resource.TestCheckResourceAttr("graylog_view.test", "title", title),
					resource.TestCheckResourceAttrSet("graylog_view.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_view.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccViewResourceConfig(updatedTitle, searchID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_view.test", "title", updatedTitle),
				),
			},
		},
	})
}

func TestAccDashboardResource(t *testing.T) {
	searchID := testAccResolveSearchID(t)
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "TF Dashboard " + suffix[:8]
	updatedTitle := title + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardResourceConfig(title, searchID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_dashboard.test", "type", "DASHBOARD"),
					resource.TestCheckResourceAttr("graylog_dashboard.test", "title", title),
					resource.TestCheckResourceAttrSet("graylog_dashboard.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDashboardResourceConfig(updatedTitle, searchID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_dashboard.test", "title", updatedTitle),
				),
			},
		},
	})
}

func testAccViewResourceConfig(title, searchID string) string {
	return fmt.Sprintf(`
resource "graylog_view" "test" {
  title       = %[1]q
  summary     = "Terraform acceptance view"
  description = "Terraform acceptance view"
  search_id   = %[2]q
  properties_json = jsonencode([])
  requires_json   = jsonencode({})
  state_json  = jsonencode({})
}
`, title, searchID)
}

func testAccDashboardResourceConfig(title, searchID string) string {
	return fmt.Sprintf(`
resource "graylog_dashboard" "test" {
  title       = %[1]q
  summary     = "Terraform acceptance dashboard"
  description = "Terraform acceptance dashboard"
  search_id   = %[2]q
  properties_json = jsonencode([])
  requires_json   = jsonencode({})
  state_json  = jsonencode({})
}
`, title, searchID)
}

func testAccResolveSearchID(t *testing.T) string {
	t.Helper()
	if fromEnv := os.Getenv("GRAYLOG_VIEW_SEARCH_ID"); fromEnv != "" {
		return fromEnv
	}

	endpoint := os.Getenv("GRAYLOG_ENDPOINT")
	username := os.Getenv("GRAYLOG_USERNAME")
	password := os.Getenv("GRAYLOG_PASSWORD")
	if endpoint == "" {
		t.Skip("GRAYLOG_ENDPOINT must be set")
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(endpoint, "/")+"/views", nil)
	if err != nil {
		t.Skipf("failed to create request for views list: %v", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-By", "terraform-provider-graylog")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("failed to query views list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Skipf("views list returned status %d", resp.StatusCode)
	}

	var result struct {
		Views []struct {
			SearchID string `json:"search_id"`
		} `json:"views"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Skipf("failed to decode views list: %v", err)
	}
	for _, v := range result.Views {
		if v.SearchID != "" {
			return v.SearchID
		}
	}

	t.Skip("no existing search_id found in Graylog views; set GRAYLOG_VIEW_SEARCH_ID to run tests")
	return ""
}
