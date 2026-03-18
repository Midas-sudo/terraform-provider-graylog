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
					resource.TestCheckResourceAttr("graylog_output.test", "type", "org.graylog2.outputs.LoggingOutput"),
					resource.TestCheckResourceAttrSet("graylog_output.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_output.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"payload_json",
				},
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
				ImportStateVerifyIgnore: []string{
					"payload_json",
				},
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
				ImportStateVerifyIgnore: []string{
					"payload_json",
				},
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

func testAccOutputResourceConfig(title string) string {
	return fmt.Sprintf(`
resource "graylog_output" "test" {
  payload_json = jsonencode({
    title = %[1]q
    type  = "org.graylog2.outputs.LoggingOutput"
    configuration = {
      prefix = "terraform-output:"
    }
  })
}
`, title)
}

func testAccExtractorResourceConfig(title string) string {
	return fmt.Sprintf(`
data "graylog_inputs" "all" {}

resource "graylog_extractor" "test" {
  input_id = data.graylog_inputs.all.inputs[0].id
  payload_json = jsonencode({
    title           = %[1]q
    source_field    = "message"
    target_field    = "message_copy"
    extractor_type  = "copy_input"
    cursor_strategy = "copy"
    condition_type  = "none"
    condition_value = ""
    extractor_config = {}
    converters       = []
    order            = 0
  })
}
`, title)
}

func testAccGrokPatternResourceConfig(name, pattern string) string {
	return fmt.Sprintf(`
resource "graylog_grok_pattern" "test" {
  payload_json = jsonencode({
    name    = %[1]q
    pattern = %[2]q
  })
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
