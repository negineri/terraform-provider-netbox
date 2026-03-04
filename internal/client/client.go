// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// NetboxClient はリトライ対応の HTTP クライアントです。
type NetboxClient struct {
	retryClient *retryablehttp.Client
	baseURL     string
	token       string
}

// NewNetboxClient はトークンを受け取り、認証付きのクライアントを生成します。
func NewNetboxClient(serverURL string, keyV2 string, tokenV2 string) *NetboxClient {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 5
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 60 * time.Second
	rc.Logger = nil
	// per-attempt タイムアウト。StandardClient を使わないためリトライ全体には影響しない。
	rc.HTTPClient.Timeout = 30 * time.Second

	// 429 と接続タイムアウトもリトライ対象にする
	rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) && urlErr.Timeout() {
				return true, nil
			}
		}
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return true, nil
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	// 429 の場合は Retry-After ヘッダーがあればその値を待機時間として使用する
	rc.Backoff = func(minWait, maxWait time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					d := time.Duration(seconds) * time.Second
					if d > maxWait {
						return maxWait
					}
					if d > 0 {
						return d
					}
				}
			}
			return maxWait
		}
		return retryablehttp.LinearJitterBackoff(minWait, maxWait, attemptNum, resp)
	}

	return &NetboxClient{
		retryClient: rc,
		baseURL:     serverURL,
		token:       "nbt_" + keyV2 + "." + tokenV2,
	}
}

func (c *NetboxClient) newRequest(ctx context.Context, method, path string, body interface{}) (*retryablehttp.Request, error) {
	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to construct full URL: %w", err)
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "terraform-provider-netbox/1.0.0")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *NetboxClient) Get(ctx context.Context, path string) (*string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.retryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	bodyString := string(bodyBytes)
	return &bodyString, nil
}

func (c *NetboxClient) Post(ctx context.Context, path string, body io.Reader) (*string, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

func (c *NetboxClient) Patch(ctx context.Context, path string, body io.Reader) (*string, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

func (c *NetboxClient) Delete(ctx context.Context, path string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *NetboxClient) doRequest(ctx context.Context, method, path string, body io.Reader) (*string, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.retryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyString := string(bodyBytes)
	return &bodyString, nil
}
