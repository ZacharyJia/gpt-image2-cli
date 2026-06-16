// Package api provides a minimal OpenAI Images API client for generations and edits.
package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultBaseURL is the OpenAI API base URL used when none is supplied.
	DefaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-image-2"
)

// Size shortcuts map user-friendly names to OpenAI size literals.
var SizeShortcuts = map[string]string{
	"1k":        "1024x1024",
	"2k":        "2048x2048",
	"4k":        "3840x2160",
	"portrait":  "1024x1536",
	"landscape": "1536x1024",
	"square":    "1024x1024",
	"wide":      "2048x1152",
	"tall":      "2160x3840",
}

// Request holds parameters shared by generate and edit calls.
type Request struct {
	Model             string
	Prompt            string
	Size              string
	Quality           string
	N                 int
	Background        string
	Moderation        string
	OutputFormat      string
	OutputCompression *int
	User              string

	// Edit-only fields.
	Images        []string // file paths
	Mask          string   // file path
	InputFidelity string
}

// Image represents one returned image.
type Image struct {
	B64JSON string `json:"b64_json"`
	URL     string `json:"url"`
}

// Response mirrors the OpenAI images response shape.
type Response struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

// Client is a tiny OpenAI Images API client.
type Client struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// New creates a Client with the provided API key and optional base URL.
// An empty baseURL falls back to DefaultBaseURL.
func New(apiKey, baseURL string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key not set")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Minute},
	}, nil
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return DefaultBaseURL
}

func (c *Client) resolveSize(size string) string {
	if v, ok := SizeShortcuts[size]; ok {
		return v
	}
	return size
}

// Generate calls POST /v1/images/generations.
func (c *Client) Generate(ctx context.Context, req Request) (*Response, error) {
	body := map[string]any{
		"model":  req.Model,
		"prompt": req.Prompt,
	}
	if size := c.resolveSize(req.Size); size != "" {
		body["size"] = size
	}
	if req.Quality != "" {
		body["quality"] = req.Quality
	}
	if req.N > 0 {
		body["n"] = req.N
	}
	if req.Background != "" {
		body["background"] = req.Background
	}
	if req.Moderation != "" {
		body["moderation"] = req.Moderation
	}
	if req.OutputFormat != "" {
		body["output_format"] = req.OutputFormat
	}
	if req.OutputCompression != nil {
		body["output_compression"] = *req.OutputCompression
	}
	if req.User != "" {
		body["user"] = req.User
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL()+"/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	return c.do(httpReq)
}

// Edit calls POST /v1/images/edits (multipart/form-data).
func (c *Client) Edit(ctx context.Context, req Request) (*Response, error) {
	for _, p := range req.Images {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("image not found: %s", p)
		}
	}
	if req.Mask != "" {
		if _, err := os.Stat(req.Mask); err != nil {
			return nil, fmt.Errorf("mask not found: %s", req.Mask)
		}
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	_ = mw.WriteField("model", req.Model)
	_ = mw.WriteField("prompt", req.Prompt)
	if size := c.resolveSize(req.Size); size != "" {
		_ = mw.WriteField("size", size)
	}
	if req.Quality != "" {
		_ = mw.WriteField("quality", req.Quality)
	}
	if req.N > 0 {
		_ = mw.WriteField("n", fmt.Sprintf("%d", req.N))
	}
	if req.Background != "" {
		_ = mw.WriteField("background", req.Background)
	}
	if req.OutputFormat != "" {
		_ = mw.WriteField("output_format", req.OutputFormat)
	}
	if req.OutputCompression != nil {
		_ = mw.WriteField("output_compression", fmt.Sprintf("%d", *req.OutputCompression))
	}
	if req.User != "" {
		_ = mw.WriteField("user", req.User)
	}

	// gpt-image-2 rejects input_fidelity; drop it locally.
	if req.InputFidelity != "" && !startsWithIgnoreCase(req.Model, "gpt-image-2") {
		_ = mw.WriteField("input_fidelity", req.InputFidelity)
	}

	for _, p := range req.Images {
		if err := writeFileField(mw, "image", p); err != nil {
			return nil, err
		}
	}
	if req.Mask != "" {
		if err := writeFileField(mw, "mask", req.Mask); err != nil {
			return nil, err
		}
	}

	if err := mw.Close(); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL()+"/images/edits", &buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())

	return c.do(httpReq)
}

func writeFileField(mw *multipart.Writer, field, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := mw.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func (c *Client) do(req *http.Request) (*Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(data))
	}
	return &result, nil
}

// FetchImage downloads an image from a URL or decodes a base64 payload.
func FetchImage(img Image, timeout time.Duration) ([]byte, error) {
	if img.B64JSON != "" {
		return base64.StdEncoding.DecodeString(img.B64JSON)
	}
	if img.URL != "" {
		client := &http.Client{Timeout: timeout}
		resp, err := client.Get(img.URL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download image: %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("image has neither b64_json nor url")
}

func startsWithIgnoreCase(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		c1, c2 := s[i], prefix[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
