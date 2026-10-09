// Package cloud holds server-only access to private Supabase services.
package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

type HTTPError struct{ Status int }

func (e *HTTPError) Error() string { return fmt.Sprintf("Supabase returned status %d", e.Status) }

func Enabled() bool { return os.Getenv("SUPABASE_URL") != "" }

func Validate() error {
	if !Enabled() && os.Getenv("VERCEL") == "" {
		return nil
	}
	endpoint, err := url.Parse(os.Getenv("SUPABASE_URL"))
	if err != nil || endpoint.Host == "" || endpoint.Scheme != "https" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || os.Getenv("SUPABASE_SERVICE_ROLE_KEY") == "" {
		return errors.New("SUPABASE_URL (HTTPS) and a server-only SUPABASE_SERVICE_ROLE_KEY are required")
	}
	return nil
}

func Request(ctx context.Context, method, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")+path, body)
	if err != nil {
		return nil, err
	}
	key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Supabase request failed")
	}
	if response.StatusCode >= 300 {
		defer response.Body.Close()
		status := response.StatusCode
		// Storage may wrap a missing bucket/object (404) in an HTTP 400 response.
		var detail struct {
			StatusCode string `json:"statusCode"`
		}
		if status == http.StatusBadRequest && json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&detail) == nil && detail.StatusCode == "404" {
			status = http.StatusNotFound
		}
		return nil, &HTTPError{Status: status}
	}
	return response, nil
}

func bucket() string {
	if name := os.Getenv("SUPABASE_STORAGE_BUCKET"); name != "" {
		return name
	}
	return "loop-media"
}

func objectPath(name string) string {
	return "/storage/v1/object/" + url.PathEscape(bucket()) + "/" + url.PathEscape(name)
}

func Upload(ctx context.Context, name, contentType string, body io.Reader) error {
	response, err := Request(ctx, http.MethodPost, objectPath(name), contentType, body)
	if err != nil {
		return err
	}
	response.Body.Close()
	return nil
}

func Delete(ctx context.Context, name string) error {
	response, err := Request(ctx, http.MethodDelete, objectPath(name), "", nil)
	if err != nil {
		return err
	}
	response.Body.Close()
	return nil
}

// Access checks happen in the Go API before reading private bucket objects.
func Read(ctx context.Context, name string) (*http.Response, error) {
	return Request(ctx, http.MethodGet, "/storage/v1/object/authenticated/"+url.PathEscape(bucket())+"/"+url.PathEscape(name), "", nil)
}

// Provision a private bucket once, without changing an existing bucket's policy.
func EnsurePrivateStorage(ctx context.Context) error {
	path := "/storage/v1/bucket/" + url.PathEscape(bucket())
	response, err := Request(ctx, http.MethodGet, path, "", nil)
	var status *HTTPError
	if errors.As(err, &status) && status.Status == 404 {
		body, _ := json.Marshal(map[string]any{"id": bucket(), "name": bucket(), "public": false, "file_size_limit": 5 * 1024 * 1024, "allowed_mime_types": []string{"image/jpeg", "image/png", "image/gif"}})
		created, createErr := Request(ctx, http.MethodPost, "/storage/v1/bucket", "application/json", bytes.NewReader(body))
		if createErr != nil {
			return createErr
		}
		created.Body.Close()
		response, err = Request(ctx, http.MethodGet, path, "", nil)
	}
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var config struct {
		ID     string `json:"id"`
		Public bool   `json:"public"`
	}
	if json.NewDecoder(response.Body).Decode(&config) != nil || config.ID != bucket() || config.Public {
		return errors.New("Supabase media bucket must exist and be private")
	}
	return nil
}
