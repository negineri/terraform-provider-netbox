package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// カスタムTransportの定義
type NetboxTransport struct {
	Token string
	// ラップする元のTransport（実際の通信を行う部分）
	Transport http.RoundTripper
}

func (t *NetboxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 【鉄則】引数で渡された元のリクエストを直接書き換えてはいけないため、クローンを作成する
	clonedReq := req.Clone(req.Context())

	// トークンをヘッダーにセット（Bearer認証の例）
	clonedReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.Token))

	// TerraformプロバイダーではUser-Agentの指定も推奨されています
	clonedReq.Header.Set("User-Agent", "terraform-provider-netbox/1.0.0")
	clonedReq.Header.Set("Content-Type", "application/json")
	clonedReq.Header.Set("Accept", "application/json")

	// 元のTransportに処理を委譲（ここで実際の通信が行われる）
	return t.Transport.RoundTrip(clonedReq)
}

// NetboxClient の定義
type NetboxClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewNetboxClient はトークンを受け取り、認証付きのクライアントを生成する
func NewNetboxClient(serverURL string, keyV2 string, tokenV2 string) *NetboxClient {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.Logger = nil // TFのログ出力に任せる

	standardClient := retryClient.StandardClient()
	standardClient.Timeout = 10 * time.Second

	// 2. Transportのラップ
	// standardClientのデフォルトTransportを、自作のAuthTransportで包む
	// 元のTransportがnilの場合は http.DefaultTransport を使うのが安全です
	baseTransport := standardClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	standardClient.Transport = &NetboxTransport{
		Token:     "nbt_" + keyV2 + "." + tokenV2,
		Transport: baseTransport,
	}

	return &NetboxClient{
		httpClient: standardClient,
		baseURL:    serverURL,
	}
}

func (c *NetboxClient) Get(ctx context.Context, path string) (*string, error) {
	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to construct full URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
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
