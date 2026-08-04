// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOutputResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "tf-output-" + suffix[:8]
	updatedTitle := title + "-upd"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutputResourceConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_output.test", "title", title),
					resource.TestCheckResourceAttr("graylog_output.test", "type", "LoggingOutput"),
					resource.TestCheckResourceAttrSet("graylog_output.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_output.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOutputResourceConfig(updatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_output.test", "title", updatedTitle),
				),
			},
		},
	})
}

func TestAccExtractorResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "tf-extractor-" + suffix[:8]
	updatedTitle := title + "-upd"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExtractorResourceConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_extractor.test", "title", title),
					resource.TestCheckResourceAttr("graylog_extractor.test", "extractor_type", "copy_input"),
					resource.TestCheckResourceAttrSet("graylog_extractor.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_extractor.test",
				ImportState:       true,
				ImportStateIdFunc: testAccExtractorImportIDFunc("graylog_extractor.test"),
				ImportStateVerify: true,
			},
			{
				Config: testAccExtractorResourceConfig(updatedTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_extractor.test", "title", updatedTitle),
				),
			},
		},
	})
}

func TestAccGrokPatternResource(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	name := "TFPATTERN" + suffix[:8]
	updatedName := name + "UPD"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGrokPatternResourceConfig(name, "foo(?<bar>.*)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_grok_pattern.test", "name", name),
					resource.TestCheckResourceAttrSet("graylog_grok_pattern.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_grok_pattern.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGrokPatternResourceConfig(updatedName, "foo(?<baz>.*)"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_grok_pattern.test", "name", updatedName),
					resource.TestCheckResourceAttr("graylog_grok_pattern.test", "pattern", "foo(?<baz>.*)"),
				),
			},
		},
	})
}

func TestAccOutputDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "tf-output-ds-" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutputDataSourcesConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_output.test", "id", "graylog_output.test", "id"),
					resource.TestCheckResourceAttr("data.graylog_output.test", "title", title),
					resource.TestCheckResourceAttrSet("data.graylog_outputs.test", "outputs.0.id"),
				),
			},
		},
	})
}

func TestAccExtractorDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	title := "tf-extractor-ds-" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExtractorDataSourcesConfig(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.graylog_extractor.test", "id", "graylog_extractor.test", "id"),
					resource.TestCheckResourceAttr("data.graylog_extractor.test", "title", title),
					resource.TestCheckResourceAttrSet("data.graylog_extractors.test", "extractors.0.id"),
				),
			},
		},
	})
}

func TestAccGrokPatternDataSources(t *testing.T) {
	suffix := strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))
	name := "TFGROKDS" + suffix[:8]

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGrokPatternDataSourcesConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.graylog_grok_patterns.test", "patterns.0.id"),
					resource.TestCheckResourceAttrSet("data.graylog_grok_patterns.test", "patterns.0.name"),
				),
			},
		},
	})
}

func testAccOutputResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_output" "test" {
  title = %[1]q
  type  = "LoggingOutput"

  configuration = {
    prefix = "terraform-output:"
  }
}
`, title)
}

func testAccOutputDataSourcesConfig(title string) string {
	return fmt.Sprintf(`
%s

data "graylog_output" "test" {
  id = graylog_output.test.id
}

data "graylog_outputs" "test" {}
`, testAccOutputResourceConfig(title))
}

func testAccExtractorDataSourcesConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_input" "extractor_ds" {
  title  = "TF Acc Extractor DS Input"
  type   = "SyslogUDPInput"
  global = true

  configuration = {
    bind_address     = "0.0.0.0"
    port             = 15141
    recv_buffer_size = 262144
  }
}

resource "graylog_extractor" "test" {
  input_id        = graylog_input.extractor_ds.id
  title           = %[1]q
  source_field    = "message"
  target_field    = "message_copy"
  extractor_type  = "copy_input"
  cursor_strategy = "copy"
  condition_type  = "none"
  condition_value = ""
  extractor_config = {}
  converters       = []
  order                 = 0
}

data "graylog_extractor" "test" {
  id       = graylog_extractor.test.id
  input_id = graylog_extractor.test.input_id
}

data "graylog_extractors" "test" {
  input_id = graylog_extractor.test.input_id
}
`, title)
}

func testAccGrokPatternDataSourcesConfig(name string) string {
	return fmt.Sprintf(`
%s

data "graylog_grok_patterns" "test" {}
`, testAccGrokPatternResourceConfig(name, "foo(?<bar>.*)"))
}

func testAccExtractorResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_input" "extractor_host" {
  title  = "TF Acc Extractor Host Input"
  type   = "SyslogUDPInput"
  global = true

  configuration = {
    bind_address     = "0.0.0.0"
    port             = 15143
    recv_buffer_size = 262144
  }
}

resource "graylog_extractor" "test" {
  input_id        = graylog_input.extractor_host.id
  title           = %[1]q
  source_field    = "message"
  target_field    = "message_copy"
  extractor_type  = "copy_input"
  cursor_strategy = "copy"
  condition_type  = "none"
  condition_value = ""
  extractor_config = {}
  converters       = []
  order                 = 0
}
`, title)
}

func testAccGrokPatternResourceConfig(name, pattern string) string {
	return fmt.Sprintf(`
resource "graylog_grok_pattern" "test" {
  name    = %[1]q
  pattern = %[2]q
}
`, name, pattern)
}

func testAccExtractorImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		inputID := rs.Primary.Attributes["input_id"]
		extractorID := rs.Primary.ID
		if inputID == "" || extractorID == "" {
			return "", fmt.Errorf("missing input_id or extractor id in state")
		}
		return inputID + "/" + extractorID, nil
	}
}
