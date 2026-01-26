package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client is a small HTTP client for Active24 API
type Client struct {
	baseURL    *neturl.URL
	httpClient *http.Client
	apiKey     string
	apiSecret  string
}

func NewClient(baseURL string, apiKey string, apiSecret string) (*Client, error) {
	parsed, err := neturl.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	hc := &http.Client{Timeout: 15 * time.Second}

	return &Client{baseURL: parsed, httpClient: hc, apiKey: apiKey, apiSecret: apiSecret}, nil
}

func (c *Client) buildURL(elem ...string) string {
	u := *c.baseURL
	// Join paths while preserving base path
	u.Path = path.Join(append([]string{c.baseURL.Path}, elem...)...)
	return u.String()
}

func (c *Client) do(ctx context.Context, method string, requestURL string, in any, out any) error {
	var body io.Reader
	var payload []byte
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = b
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	// Active24 v2 HMAC Basic: password is signature of canonical request
	now := time.Now().UTC()
	unixTs := strconv.FormatInt(now.Unix(), 10)
	parsedURL, _ := neturl.Parse(requestURL)
	// Sign ONLY the path per observed behavior (omit query from canonical)
	canonical := fmt.Sprintf("%s %s %s", method, parsedURL.Path, unixTs)
	mac := hmac.New(sha1.New, []byte(c.apiSecret))
	mac.Write([]byte(canonical))
	signature := fmt.Sprintf("%x", mac.Sum(nil))
	auth := base64.StdEncoding.EncodeToString([]byte(c.apiKey + ":" + signature))

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("User-Agent", "terraform-provider-active24")
	req.Header.Set("Accept", "application/json")
	// X-Date must match the timestamp used in the canonical string
	req.Header.Set("X-Date", now.Format("20060102T150405Z"))
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isDebugEnabled() {
			fmt.Printf("[active24] request error %s %s: %v\n", method, requestURL, err)
		}
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if isDebugEnabled() {
		fmt.Printf("[active24] %s %s -> %s\n", method, requestURL, resp.Status)
		if len(payload) > 0 {
			fmt.Printf("[active24] request body: %s\n", string(payload))
		}
		fmt.Printf("[active24] response body: %s\n", string(respBytes))
		// Also log via Terraform logger so it shows with TF_LOG
		tflog.Info(ctx, "active24 http", map[string]any{
			"method":        method,
			"url":           requestURL,
			"status":        resp.Status,
			"request_body":  string(payload),
			"response_body": string(respBytes),
		})
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("active24 API error: %s: %s", resp.Status, string(respBytes))
	}

	if out != nil && len(respBytes) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(respBytes))
		return decoder.Decode(out)
	}

	return nil
}

// DNS record models (based on common DNS fields; may need adjustments per API)
type DNSRecord struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Content  string  `json:"content"`
	TTL      int64   `json:"ttl"`
	Priority *int64  `json:"priority,omitempty"`
	CAAValue string  `json:"caaValue,omitempty"`
	Flags    *int64  `json:"flags,omitempty"`
	Tag      string  `json:"tag,omitempty"`
}

type createRecordRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Content  string  `json:"content"`
	TTL      int64   `json:"ttl"`
	Priority *int64  `json:"priority,omitempty"`
	CAAValue string  `json:"caaValue,omitempty"`
	Flags    *int64  `json:"flags,omitempty"`
	Tag      string  `json:"tag,omitempty"`
}

// CreateRecord creates a DNS record under a domain
func (c *Client) CreateRecord(ctx context.Context, domain string, req createRecordRequest) (*DNSRecord, error) {
	var out DNSRecord
	// v2 API path: /v2/service/{service}/dns/record
	url := c.buildURL("service", domain, "dns", "record")
	if err := c.do(ctx, http.MethodPost, url, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetRecord(ctx context.Context, domain string, id int64) (*DNSRecord, error) {
	var out DNSRecord
	url := c.buildURL("service", domain, "dns", "record", fmt.Sprintf("%d", id))
	if err := c.do(ctx, http.MethodGet, url, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateRecord(ctx context.Context, domain string, id int64, req createRecordRequest) (*DNSRecord, error) {
	var out DNSRecord
	url := c.buildURL("service", domain, "dns", "record", fmt.Sprintf("%d", id))
	if err := c.do(ctx, http.MethodPut, url, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteRecord(ctx context.Context, domain string, id int64) error {
	url := c.buildURL("service", domain, "dns", "record", fmt.Sprintf("%d", id))
	return c.do(ctx, http.MethodDelete, url, nil, nil)
}

// dnsRecordsPage is a minimal response model for paginated list
type dnsRecordsPage struct {
	Data []DNSRecord `json:"data"`
}

// ListRecords lists DNS records with basic filters (name/type/content/ttl)
func (c *Client) ListRecords(ctx context.Context, domain string, name string, rtype string, content string, ttl *int64) ([]DNSRecord, error) {
	base := c.buildURL("service", domain, "dns", "record")
	q := neturl.Values{}
	if name != "" {
		q.Set("filters[name]", name)
	}
	if rtype != "" {
		q.Set("filters[type]", rtype)
	}
	if content != "" {
		q.Set("filters[content]", content)
	}
	if ttl != nil {
		q.Set("filters[ttl]", fmt.Sprintf("%d", *ttl))
	}
	if enc := q.Encode(); enc != "" {
		base = base + "?" + enc
	}

	var page dnsRecordsPage
	if err := c.do(ctx, http.MethodGet, base, nil, &page); err != nil {
		return nil, err
	}
	return page.Data, nil
}

func getEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return ""
}

func isDebugEnabled() bool {
	v := getEnv("ACTIVE24_DEBUG")
	if v == "1" || v == "true" || v == "TRUE" || v == "yes" || v == "on" {
		return true
	}
	return false
}
