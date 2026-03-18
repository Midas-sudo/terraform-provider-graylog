package provider

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

func contentPackFromPayload(payload string) (*client.ContentPack, diag.Diagnostics) {
	var diags diag.Diagnostics
	var contentPack client.ContentPack
	if err := json.Unmarshal([]byte(payload), &contentPack); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &contentPack, diags
}

func contentPackInstallationFromPayload(payload string) (*client.ContentPackInstallationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var req client.ContentPackInstallationRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		diags.AddError("Invalid payload_json", fmt.Sprintf("Failed to parse payload_json: %v", err))
	}
	return &req, diags
}

func marshalContentPackJSON(contentPack *client.ContentPack) string {
	b, err := json.Marshal(contentPack)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func marshalContentPackInstallationJSON(req *client.ContentPackInstallationRequest) string {
	b, err := json.Marshal(req)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func mapContentPackToResourceModel(contentPack *client.ContentPack, data *ContentPackResourceModel) {
	data.ContentPackID = types.StringValue(contentPack.ID)
	data.Revision = types.Int64Value(contentPack.Rev)
	data.Name = types.StringValue(contentPack.Name)
	data.ID = types.StringValue(fmt.Sprintf("%s/%d", contentPack.ID, contentPack.Rev))
}

func mapContentPackInstallationToResourceModel(inst *client.ContentPackInstallation, data *ContentPackInstallationResourceModel) {
	data.ID = types.StringValue(inst.ID)
	data.ContentPackID = types.StringValue(inst.ContentPackID)
	data.Revision = types.Int64Value(inst.ContentPackRevision)
}

func parseContentPackImportID(importID string) (string, int64, error) {
	parts := strings.Split(importID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("invalid import identifier, expected content_pack_id/revision")
	}
	rev, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid revision %q: %w", parts[1], err)
	}
	return parts[0], rev, nil
}
