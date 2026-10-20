// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the HASHICUPS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "netbox" {
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"netbox": providerserver.NewProtocol6WithError(New("test")()),
	}
)

// testAccClient は環境変数から acceptance test 用の client を作ります。
func testAccClient() *client.NetboxClient {
	if token := os.Getenv("NETBOX_TOKEN"); token != "" {
		return client.NewNetboxClientV1(os.Getenv("NETBOX_SERVER_URL"), token)
	}
	return client.NewNetboxClient(os.Getenv("NETBOX_SERVER_URL"), os.Getenv("NETBOX_KEY_V2"), os.Getenv("NETBOX_TOKEN_V2"))
}

// testAccSkipIfNetboxOlderThan は接続先 NetBox が major.minor より古い場合にテストをスキップします。
func testAccSkipIfNetboxOlderThan(t *testing.T, major, minor int) {
	t.Helper()

	bodyStr, err := testAccClient().Get(context.Background(), "api/status/")
	if err != nil {
		t.Fatalf("failed to get NetBox status: %s", err)
	}
	var status struct {
		NetboxVersion string `json:"netbox-version"`
	}
	if err := json.Unmarshal([]byte(*bodyStr), &status); err != nil {
		t.Fatalf("failed to parse NetBox status: %s", err)
	}

	var gotMajor, gotMinor int
	if _, err := fmt.Sscanf(status.NetboxVersion, "%d.%d", &gotMajor, &gotMinor); err != nil {
		t.Fatalf("failed to parse NetBox version %q: %s", status.NetboxVersion, err)
	}
	if gotMajor < major || (gotMajor == major && gotMinor < minor) {
		t.Skipf("requires NetBox %d.%d or later, got %s", major, minor, status.NetboxVersion)
	}
}
