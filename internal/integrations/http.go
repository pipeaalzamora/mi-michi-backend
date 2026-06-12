package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 12 * time.Second}

func getJSON(ctx context.Context, endpoint string, headers map[string]string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent())
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = res.Status
		}
		return fmt.Errorf("API externa respondió %s: %s", res.Status, msg)
	}

	return json.NewDecoder(res.Body).Decode(target)
}

func postFormJSON(ctx context.Context, endpoint string, form url.Values, headers map[string]string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", defaultUserAgent())
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = res.Status
		}
		return fmt.Errorf("API externa respondió %s: %s", res.Status, msg)
	}

	return json.NewDecoder(res.Body).Decode(target)
}

func defaultUserAgent() string {
	return "MiMichi/1.0 (configura-contacto@example.com)"
}
