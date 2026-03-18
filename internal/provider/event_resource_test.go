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

func TestAccEventNotificationResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "TF Event Notification " + suffix[:8]
	updatedTitle := title + " Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventNotificationResourceConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_event_notification.test", "title", title),
					resource.TestCheckResourceAttrSet("graylog_event_notification.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_event_notification.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"config_json",
				},
			},
			{
				Config: testAccEventNotificationResourceConfig(updatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_event_notification.test", "title", updatedTitle),
				),
			},
		},
	})
}

func TestAccEventDefinitionResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "TF Event Definition " + suffix[:8]
	updatedTitle := title + " Updated"
	storageStreamID := testAccResolveEventStorageStreamID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventDefinitionResourceConfig(title, storageStreamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_event_definition.test", "title", title),
					resource.TestCheckResourceAttr("graylog_event_definition.test", "state", "ENABLED"),
					resource.TestCheckResourceAttrSet("graylog_event_definition.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_event_definition.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventDefinitionResourceConfig(updatedTitle, storageStreamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_event_definition.test", "title", updatedTitle),
				),
			},
		},
	})
}

func TestAccEventDefinitionNotificationBindingResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	storageStreamID := testAccResolveEventStorageStreamID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventBindingResourceConfig(suffix[:8], storageStreamID, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("graylog_event_definition_notification_binding.test", "id"),
					resource.TestCheckResourceAttr("graylog_event_definition_notification_binding.test", "notification_ids.#", "1"),
				),
			},
			{
				ResourceName:      "graylog_event_definition_notification_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventBindingResourceConfig(suffix[:8], storageStreamID, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_event_definition_notification_binding.test", "notification_ids.#", "2"),
				),
			},
		},
	})
}

func testAccEventNotificationResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_event_notification" "test" {
  title       = %[1]q
  description = "Terraform acceptance event notification"
  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/terraform-event-notification"
  })
}
`, title)
}

func testAccEventDefinitionResourceConfig(title, storageStreamID string) string {
	return fmt.Sprintf(`
resource "graylog_event_definition" "test" {
  title       = %[1]q
  description = "Terraform acceptance event definition"
  priority    = 1
  alert       = false
  config_json = jsonencode({
    type = "system-notifications-v1"
  })
  field_spec_json = jsonencode({})
  key_spec        = []
  notification_settings_json = jsonencode({
    grace_period_ms = 0
    backlog_size    = 0
  })
  notifications_json = jsonencode([])
  storage_json = jsonencode([
    {
      type    = "persist-to-streams-v1"
      streams = [%[2]q]
    }
  ])
  state = "ENABLED"
}
`, title, storageStreamID)
}

func testAccEventBindingResourceConfig(suffix, storageStreamID string, includeSecond bool) string {
	definitionTitle := "TF Event Binding Definition " + suffix
	notificationOneTitle := "TF Event Binding Notification One " + suffix
	notificationTwoTitle := "TF Event Binding Notification Two " + suffix

	if includeSecond {
		return fmt.Sprintf(`
resource "graylog_event_notification" "one" {
  title       = %[1]q
  description = "Terraform acceptance event notification one"
  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/terraform-event-binding-one"
  })
}

resource "graylog_event_notification" "two" {
  title       = %[2]q
  description = "Terraform acceptance event notification two"
  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/terraform-event-binding-two"
  })
}

resource "graylog_event_definition" "test" {
  title       = %[3]q
  description = "Terraform acceptance event definition for binding"
  priority    = 1
  alert       = false
  config_json = jsonencode({
    type = "system-notifications-v1"
  })
  field_spec_json = jsonencode({})
  key_spec        = []
  notification_settings_json = jsonencode({
    grace_period_ms = 0
    backlog_size    = 0
  })
  notifications_json = jsonencode([])
  storage_json = jsonencode([
    {
      type    = "persist-to-streams-v1"
      streams = [%[4]q]
    }
  ])
  state = "ENABLED"
}

resource "graylog_event_definition_notification_binding" "test" {
  event_definition_id = graylog_event_definition.test.id
  notification_ids = [
    graylog_event_notification.one.id,
    graylog_event_notification.two.id,
  ]
}
`, notificationOneTitle, notificationTwoTitle, definitionTitle, storageStreamID)
	}

	return fmt.Sprintf(`
resource "graylog_event_notification" "one" {
  title       = %[1]q
  description = "Terraform acceptance event notification one"
  config_json = jsonencode({
    type = "http-notification-v1"
    url  = "https://example.org/terraform-event-binding-one"
  })
}

resource "graylog_event_definition" "test" {
  title       = %[2]q
  description = "Terraform acceptance event definition for binding"
  priority    = 1
  alert       = false
  config_json = jsonencode({
    type = "system-notifications-v1"
  })
  field_spec_json = jsonencode({})
  key_spec        = []
  notification_settings_json = jsonencode({
    grace_period_ms = 0
    backlog_size    = 0
  })
  notifications_json = jsonencode([])
  storage_json = jsonencode([
    {
      type    = "persist-to-streams-v1"
      streams = [%[3]q]
    }
  ])
  state = "ENABLED"
}

resource "graylog_event_definition_notification_binding" "test" {
  event_definition_id = graylog_event_definition.test.id
  notification_ids = [
    graylog_event_notification.one.id,
  ]
}
`, notificationOneTitle, definitionTitle, storageStreamID)
}

func testAccResolveEventStorageStreamID(t *testing.T) string {
	t.Helper()
	if fromEnv := os.Getenv("GRAYLOG_EVENT_STORAGE_STREAM_ID"); fromEnv != "" {
		return fromEnv
	}

	endpoint := os.Getenv("GRAYLOG_ENDPOINT")
	username := os.Getenv("GRAYLOG_USERNAME")
	password := os.Getenv("GRAYLOG_PASSWORD")
	if endpoint == "" {
		t.Skip("GRAYLOG_ENDPOINT must be set")
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(endpoint, "/")+"/events/definitions", nil)
	if err != nil {
		t.Skipf("failed to create request for event definitions list: %v", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-By", "terraform-provider-graylog")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("failed to query event definitions list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Skipf("event definitions list returned status %d", resp.StatusCode)
	}

	var result struct {
		EventDefinitions []struct {
			Config struct {
				Type string `json:"type"`
			} `json:"config"`
			Storage []struct {
				Type    string   `json:"type"`
				Streams []string `json:"streams"`
			} `json:"storage"`
		} `json:"event_definitions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Skipf("failed to decode event definitions list: %v", err)
	}

	for _, def := range result.EventDefinitions {
		if def.Config.Type != "system-notifications-v1" {
			continue
		}
		for _, storage := range def.Storage {
			if storage.Type == "persist-to-streams-v1" && len(storage.Streams) > 0 {
				return storage.Streams[0]
			}
		}
	}

	return "000000000000000000000003"
}
