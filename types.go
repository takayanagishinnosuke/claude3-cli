package main

type Request struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	System           string    `json:"system"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature"`
	TopP             float64   `json:"top_p"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type RequestImg struct {
	AnthropicVersion string       `json:"anthropic_version"`
	MaxTokens        int          `json:"max_tokens"`
	System           string       `json:"system"`
	Messages         []MessageImg `json:"messages"`
}

type MessageImg struct {
	Role    string       `json:"role"`
	Content []ContentImg `json:"content"`
}

type ContentImg struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	Source *Source `json:"source,omitempty"`
}

type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Response struct {
	ID           string        `json:"id"`
	Model        string        `json:"model"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	ContentItem  []ContentItem `json:"content"`
	StopReason   string        `json:"stop_reason,omitempty"`   // オプショナルフィールド
	StopSequence string        `json:"stop_sequence,omitempty"` // オプショナルフィールド
	Usage        UsageDetails  `json:"usage"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type UsageDetails struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type StreamResponse struct {
	Type  string    `json:"type"`
	Index int       `json:"index"`
	Delta TextDelta `json:"delta"`
}

type TextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
