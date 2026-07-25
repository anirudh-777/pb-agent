package pocketbase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anirudh-777/pb-agent/internal/config"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type APIError struct {
	Status  int
	Message string
	Data    any
}

func (e *APIError) Error() string {
	return fmt.Sprintf("PocketBase returned %d: %s", e.Status, e.Message)
}

type ListResult struct {
	Page       int              `json:"page"`
	PerPage    int              `json:"perPage"`
	TotalItems int              `json:"totalItems"`
	TotalPages int              `json:"totalPages"`
	Items      []map[string]any `json:"items"`
}

func New(connection config.Connection, token string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if connection.AllowInsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit development/test setting.
	}
	return &Client{
		baseURL: strings.TrimRight(connection.URL, "/"),
		token:   token,
		http:    &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}
}

func (c *Client) Fingerprint() string {
	sum := sha256.Sum256([]byte(c.baseURL))
	return hex.EncodeToString(sum[:8])
}

func (c *Client) Request(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		request.Header.Set("Authorization", c.token)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 400 {
		apiErr := &APIError{Status: response.StatusCode, Message: http.StatusText(response.StatusCode)}
		var decoded struct {
			Message string `json:"message"`
			Data    any    `json:"data"`
		}
		if json.Unmarshal(raw, &decoded) == nil {
			if decoded.Message != "" {
				apiErr.Message = decoded.Message
			}
			apiErr.Data = decoded.Data
		}
		return nil, apiErr
	}
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	return raw, nil
}

func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	raw, err := c.Request(ctx, http.MethodGet, "/api/health", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, json.Unmarshal(raw, &result)
}

func (c *Client) Collections(ctx context.Context, page, perPage int) (ListResult, error) {
	path := "/api/collections?page=" + strconv.Itoa(page) + "&perPage=" + strconv.Itoa(perPage)
	raw, err := c.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListResult{}, err
	}
	var result ListResult
	return result, json.Unmarshal(raw, &result)
}

func (c *Client) Collection(ctx context.Context, nameOrID string) (map[string]any, error) {
	raw, err := c.Request(ctx, http.MethodGet, "/api/collections/"+url.PathEscape(nameOrID), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, json.Unmarshal(raw, &result)
}

func (c *Client) Records(ctx context.Context, collection string, page, perPage int, filter, sort string) (ListResult, error) {
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("perPage", strconv.Itoa(perPage))
	if filter != "" {
		query.Set("filter", filter)
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	path := "/api/collections/" + url.PathEscape(collection) + "/records?" + query.Encode()
	raw, err := c.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ListResult{}, err
	}
	var result ListResult
	return result, json.Unmarshal(raw, &result)
}

func (c *Client) Record(ctx context.Context, collection, id string) (map[string]any, error) {
	path := "/api/collections/" + url.PathEscape(collection) + "/records/" + url.PathEscape(id)
	raw, err := c.Request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	return result, json.Unmarshal(raw, &result)
}

func (c *Client) Logs(ctx context.Context, page, perPage int) (json.RawMessage, error) {
	path := "/api/logs?page=" + strconv.Itoa(page) + "&perPage=" + strconv.Itoa(perPage) + "&sort=-created"
	return c.Request(ctx, http.MethodGet, path, nil)
}

func (c *Client) Backups(ctx context.Context) (json.RawMessage, error) {
	return c.Request(ctx, http.MethodGet, "/api/backups", nil)
}

func (c *Client) DownloadFile(ctx context.Context, collection, record, filename string, writer io.Writer) error {
	path := "/api/files/" + url.PathEscape(collection) + "/" + url.PathEscape(record) + "/" + url.PathEscape(filename)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		request.Header.Set("Authorization", c.token)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return &APIError{Status: response.StatusCode, Message: http.StatusText(response.StatusCode)}
	}
	const maxFileSize = 256 << 20
	written, err := io.Copy(writer, io.LimitReader(response.Body, maxFileSize+1))
	if err != nil {
		return err
	}
	if written > maxFileSize {
		return fmt.Errorf("file exceeds the 256 MiB download limit")
	}
	return nil
}
