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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccContentPackResource(t *testing.T) {
	uuid := strings.ToLower(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		time.Now().UnixNano()&0xffffffff,
		(time.Now().UnixNano()>>32)&0xffff,
		(time.Now().UnixNano()>>48)&0xffff,
		(time.Now().UnixNano()>>12)&0xffff,
		(time.Now().UnixNano()>>16)&0xffffffffffff,
	))
	name := "tf-content-pack-" + uuid[:8]
	updatedName := name + "-upd"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentPackResourceConfig(uuid, 1, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_content_pack.test", "content_pack_id", uuid),
					resource.TestCheckResourceAttr("graylog_content_pack.test", "revision", "1"),
					resource.TestCheckResourceAttr("graylog_content_pack.test", "name", name),
				),
			},
			{
				ResourceName:      "graylog_content_pack.test",
				ImportState:       true,
				ImportStateIdFunc: testAccContentPackImportIDFunc("graylog_content_pack.test"),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"payload_json",
				},
			},
			{
				Config: testAccContentPackResourceConfig(uuid, 1, updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_content_pack.test", "name", updatedName),
				),
			},
		},
	})
}

func TestAccContentPackInstallationResource(t *testing.T) {
	contentPackID, revision := testAccResolveInstallableContentPack(t)
	comment := "terraform-install-" + strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano()))[:8]
	updatedComment := comment + "-upd"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContentPackInstallationResourceConfig(contentPackID, revision, comment),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_content_pack_installation.test", "content_pack_id", contentPackID),
					resource.TestCheckResourceAttr("graylog_content_pack_installation.test", "revision", fmt.Sprintf("%d", revision)),
					resource.TestCheckResourceAttrSet("graylog_content_pack_installation.test", "id"),
				),
			},
			{
				ResourceName:      "graylog_content_pack_installation.test",
				ImportState:       true,
				ImportStateIdFunc: testAccContentPackInstallationImportIDFunc("graylog_content_pack_installation.test"),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"payload_json",
				},
			},
			{
				Config: testAccContentPackInstallationResourceConfig(contentPackID, revision, updatedComment),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("graylog_content_pack_installation.test", "content_pack_id", contentPackID),
					resource.TestCheckResourceAttr("graylog_content_pack_installation.test", "revision", fmt.Sprintf("%d", revision)),
				),
			},
		},
	})
}

func testAccContentPackResourceConfig(contentPackID string, revision int64, name string) string {
	return fmt.Sprintf(`
resource "graylog_content_pack" "test" {
  payload_json = jsonencode({
    id          = %[1]q
    v           = "1"
    rev         = %[2]d
    name        = %[3]q
    summary     = "Terraform content pack"
    description = "Terraform acceptance content pack"
    vendor      = "Terraform"
    url         = "https://example.org"
    parameters  = []
    entities    = []
  })
}
`, contentPackID, revision, name)
}

func testAccContentPackInstallationResourceConfig(contentPackID string, revision int64, comment string) string {
	return fmt.Sprintf(`
resource "graylog_content_pack_installation" "test" {
  content_pack_id = %[1]q
  revision        = %[2]d
  payload_json = jsonencode({
    comment    = %[3]q
    parameters = {}
  })
}
`, contentPackID, revision, comment)
}

func testAccContentPackImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		cpID := rs.Primary.Attributes["content_pack_id"]
		rev := rs.Primary.Attributes["revision"]
		if cpID == "" || rev == "" {
			return "", fmt.Errorf("missing content_pack_id or revision in state")
		}
		return cpID + "/" + rev, nil
	}
}

func testAccContentPackInstallationImportIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		cpID := rs.Primary.Attributes["content_pack_id"]
		instID := rs.Primary.ID
		if cpID == "" || instID == "" {
			return "", fmt.Errorf("missing content_pack_id or installation id in state")
		}
		return cpID + "/" + instID, nil
	}
}

func testAccResolveInstallableContentPack(t *testing.T) (string, int64) {
	t.Helper()
	if cpID := os.Getenv("GRAYLOG_CONTENT_PACK_ID"); cpID != "" {
		rev := int64(1)
		if revEnv := os.Getenv("GRAYLOG_CONTENT_PACK_REVISION"); revEnv != "" {
			fmt.Sscanf(revEnv, "%d", &rev)
		}
		return cpID, rev
	}

	endpoint := os.Getenv("GRAYLOG_ENDPOINT")
	username := os.Getenv("GRAYLOG_USERNAME")
	password := os.Getenv("GRAYLOG_PASSWORD")
	if endpoint == "" {
		t.Skip("GRAYLOG_ENDPOINT must be set")
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(endpoint, "/")+"/system/content_packs/latest", nil)
	if err != nil {
		t.Skipf("failed to create request for content packs: %v", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-By", "terraform-provider-graylog")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("failed to query content packs list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Skipf("content packs list returned status %d", resp.StatusCode)
	}

	var result struct {
		ContentPacks []struct {
			ID  string `json:"id"`
			Rev int64  `json:"rev"`
		} `json:"content_packs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Skipf("failed to decode content packs list: %v", err)
	}
	if len(result.ContentPacks) == 0 {
		t.Skip("no content packs available to test installation")
	}
	return result.ContentPacks[0].ID, result.ContentPacks[0].Rev
}
