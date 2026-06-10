// Copyright (c) negineri 2026
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
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// 死活監視のデフォルト値。テストでは短い値に差し替える。
var (
	// healthCheckInterval は死活監視で API を叩く間隔です。
	healthCheckInterval = 30 * time.Second
	// healthCheckTimeout は死活監視 1 回あたりのタイムアウトです。
	// HTTPClient.Timeout は無効化されているため、ハングしたリクエストに巻き込まれないよう独自にタイムアウトを設定する。
	healthCheckTimeout = 10 * time.Second
	// healthCheckFailureThreshold はハングとみなして警告ログを出すまでの連続失敗回数です。
	healthCheckFailureThreshold = 3
)

// NetboxClient はリトライ対応の HTTP クライアントです。
type NetboxClient struct {
	retryClient       *retryablehttp.Client
	baseURL           string
	authHeader        string
	healthCheckCancel context.CancelFunc
}

// NewNetboxClient は v2 トークン（keyV2 + tokenV2）を受け取り、認証付きのクライアントを生成します。
func NewNetboxClient(serverURL string, keyV2 string, tokenV2 string) *NetboxClient {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 5
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 60 * time.Second
	rc.Logger = nil
	// タイムアウトを無効化し、応答があるまで無限に待機する。
	// NetBox のハングは死活監視による定期ポーリングで別途検知する。
	rc.HTTPClient.Timeout = 0

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
		authHeader:  "Bearer nbt_" + keyV2 + "." + tokenV2,
	}
}

// NewNetboxClientV1 は v1 トークンを受け取り、認証付きのクライアントを生成します。
func NewNetboxClientV1(serverURL string, token string) *NetboxClient {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 5
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 60 * time.Second
	rc.Logger = nil
	// タイムアウトを無効化し、応答があるまで無限に待機する。
	// NetBox のハングは死活監視による定期ポーリングで別途検知する。
	rc.HTTPClient.Timeout = 0

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
		authHeader:  "Token " + token,
	}
}

// StartHealthCheck は NetBox API に対する死活監視を開始します。
// HTTPClient.Timeout を無効化しているため、API がハングして応答しなくなっても
// 通常のリクエストはタイムアウトしません。本メソッドはそれとは独立に
// 短いタイムアウトで定期的に api/status/ を叩き、連続して失敗した場合に
// ハングの可能性をログへ出力します。
//
// ctx はログ出力先の設定を引き継ぐために使用しますが、ctx のキャンセルには
// 影響されず、StopHealthCheck が呼ばれるまで監視を継続します。
func (c *NetboxClient) StartHealthCheck(ctx context.Context) {
	if c.healthCheckCancel != nil {
		return
	}

	healthCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c.healthCheckCancel = cancel

	go c.healthCheckLoop(healthCtx)
}

// StopHealthCheck は StartHealthCheck で開始した死活監視を停止します。
func (c *NetboxClient) StopHealthCheck() {
	if c.healthCheckCancel == nil {
		return
	}
	c.healthCheckCancel()
	c.healthCheckCancel = nil
}

func (c *NetboxClient) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
			_, err := c.Get(checkCtx, "api/status/")
			cancel()

			if err != nil {
				consecutiveFailures++
				if consecutiveFailures >= healthCheckFailureThreshold {
					tflog.Error(ctx, "Netbox API health check failed repeatedly; the API may be hanging or unreachable", map[string]interface{}{
						"consecutive_failures": consecutiveFailures,
						"error":                err.Error(),
					})
				} else {
					tflog.Warn(ctx, "Netbox API health check failed", map[string]interface{}{
						"consecutive_failures": consecutiveFailures,
						"error":                err.Error(),
					})
				}
				continue
			}

			if consecutiveFailures >= healthCheckFailureThreshold {
				tflog.Info(ctx, "Netbox API health check recovered")
			}
			consecutiveFailures = 0
		}
	}
}

func (c *NetboxClient) newRequest(ctx context.Context, method, path string, body interface{}) (*retryablehttp.Request, error) {
	// クエリ文字列を分離してから url.JoinPath でパスを結合する。
	// url.JoinPath に "?" を含む文字列を渡すと "%3F" にエンコードされてしまうため。
	pathOnly, rawQuery, _ := strings.Cut(path, "?")
	fullURL, err := url.JoinPath(c.baseURL, pathOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to construct full URL: %w", err)
	}
	if rawQuery != "" {
		fullURL += "?" + rawQuery
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader)
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
