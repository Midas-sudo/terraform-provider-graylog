package client

import (
	"context"
	"fmt"
)

type Output struct {
	ID            string                 `json:"id,omitempty"`
	Title         string                 `json:"title"`
	Type          string                 `json:"type"`
	Configuration map[string]interface{} `json:"configuration"`
	CreatorUserID string                 `json:"creator_user_id,omitempty"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	ContentPack   *string                `json:"content_pack,omitempty"`
}

type OutputsResponse struct {
	Total   int      `json:"total"`
	Outputs []Output `json:"outputs"`
}

type createExtractorResponse struct {
	ExtractorID string `json:"extractor_id"`
}

type Extractor struct {
	ID              string                   `json:"id,omitempty"`
	Title           string                   `json:"title"`
	Type            string                   `json:"type,omitempty"`
	ExtractorType   string                   `json:"extractor_type,omitempty"`
	CursorStrategy  string                   `json:"cursor_strategy"`
	SourceField     string                   `json:"source_field"`
	TargetField     string                   `json:"target_field"`
	ExtractorConfig map[string]interface{}   `json:"extractor_config"`
	ConditionType   string                   `json:"condition_type"`
	ConditionValue  string                   `json:"condition_value"`
	Converters      []map[string]interface{} `json:"converters"`
	Order           int64                    `json:"order,omitempty"`
}

type ExtractorsResponse struct {
	Total      int         `json:"total"`
	Extractors []Extractor `json:"extractors"`
}

type GrokPattern struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Pattern     string  `json:"pattern"`
	ContentPack *string `json:"content_pack,omitempty"`
}

type GrokPatternsResponse struct {
	Patterns []GrokPattern `json:"patterns"`
}

func (c *Client) GetOutputs(ctx context.Context) (*OutputsResponse, error) {
	var result OutputsResponse
	if err := c.Get(ctx, "/system/outputs", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetOutput(ctx context.Context, id string) (*Output, error) {
	var result Output
	if err := c.Get(ctx, fmt.Sprintf("/system/outputs/%s", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateOutput(ctx context.Context, output *Output) (*Output, error) {
	var result Output
	if err := c.Post(ctx, "/system/outputs", output, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateOutput(ctx context.Context, id string, output *Output) (*Output, error) {
	var result Output
	if err := c.Put(ctx, fmt.Sprintf("/system/outputs/%s", id), output, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteOutput(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/outputs/%s", id))
}

func (c *Client) GetExtractors(ctx context.Context, inputID string) (*ExtractorsResponse, error) {
	var result ExtractorsResponse
	if err := c.Get(ctx, fmt.Sprintf("/system/inputs/%s/extractors", inputID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetExtractor(ctx context.Context, inputID, extractorID string) (*Extractor, error) {
	var result Extractor
	if err := c.Get(ctx, fmt.Sprintf("/system/inputs/%s/extractors/%s", inputID, extractorID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateExtractor(ctx context.Context, inputID string, extractor *Extractor) (*Extractor, error) {
	var created createExtractorResponse
	if err := c.Post(ctx, fmt.Sprintf("/system/inputs/%s/extractors", inputID), extractorToRequest(extractor), &created); err != nil {
		return nil, err
	}
	return c.GetExtractor(ctx, inputID, created.ExtractorID)
}

func (c *Client) UpdateExtractor(ctx context.Context, inputID, extractorID string, extractor *Extractor) (*Extractor, error) {
	var result Extractor
	if err := c.Put(ctx, fmt.Sprintf("/system/inputs/%s/extractors/%s", inputID, extractorID), extractorToRequest(extractor), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteExtractor(ctx context.Context, inputID, extractorID string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/inputs/%s/extractors/%s", inputID, extractorID))
}

func (c *Client) GetGrokPatterns(ctx context.Context) (*GrokPatternsResponse, error) {
	var result GrokPatternsResponse
	if err := c.Get(ctx, "/system/grok", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetGrokPattern(ctx context.Context, id string) (*GrokPattern, error) {
	var result GrokPattern
	if err := c.Get(ctx, fmt.Sprintf("/system/grok/%s", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CreateGrokPattern(ctx context.Context, pattern *GrokPattern) (*GrokPattern, error) {
	var result GrokPattern
	if err := c.Post(ctx, "/system/grok", pattern, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateGrokPattern(ctx context.Context, id string, pattern *GrokPattern) (*GrokPattern, error) {
	pattern.ID = id
	var result GrokPattern
	if err := c.Put(ctx, fmt.Sprintf("/system/grok/%s", id), pattern, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteGrokPattern(ctx context.Context, id string) error {
	return c.Delete(ctx, fmt.Sprintf("/system/grok/%s", id))
}

func extractorToRequest(extractor *Extractor) *Extractor {
	req := &Extractor{
		Title:           extractor.Title,
		ExtractorType:   extractor.ExtractorType,
		CursorStrategy:  extractor.CursorStrategy,
		SourceField:     extractor.SourceField,
		TargetField:     extractor.TargetField,
		ExtractorConfig: extractor.ExtractorConfig,
		ConditionType:   extractor.ConditionType,
		ConditionValue:  extractor.ConditionValue,
		Converters:      extractor.Converters,
		Order:           extractor.Order,
	}

	if req.ExtractorType == "" {
		req.ExtractorType = extractor.Type
	}
	if req.ExtractorConfig == nil {
		req.ExtractorConfig = map[string]interface{}{}
	}
	if req.Converters == nil {
		req.Converters = []map[string]interface{}{}
	}

	return req
}
