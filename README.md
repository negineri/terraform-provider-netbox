# Terraform Provider for NetBox

[NetBox](https://netboxlabs.com/) の Terraform Provider です。[Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) を使用して構築されています。

## 要件

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## 使い方

```terraform
terraform {
  required_providers {
    netbox = {
      source = "negineri/netbox"
    }
  }
}

# Configuration-based authentication
provider "netbox" {
  server_url = "https://netbox.example.com:8000"
  key_v2     = "testkey"
  token_v2   = "testtoken"
}
```

認証情報は環境変数でも設定できます。

| 環境変数               | 説明                                  |
| ---------------------- | ------------------------------------- |
| `NETBOX_SERVER_URL`    | NetBox API の URL                     |
| `NETBOX_KEY_V2`        | API key (v2 token 用)                 |
| `NETBOX_TOKEN_V2`      | API token (v2 token 用)               |
| `NETBOX_TOKEN`         | API v1 token                          |

## サポートするリソースとデータソース

### Resources

| リソース名                               | 説明                             |
| ---------------------------------------- | -------------------------------- |
| `netbox_available_ip`                    | プレフィックスから利用可能な IP アドレスを割り当て |
| `netbox_available_prefix`                | プレフィックスから利用可能なサブネットを割り当て |
| `netbox_custom_field`                    | カスタムフィールド               |
| `netbox_device`                          | デバイス                         |
| `netbox_device_interface`                | デバイスインターフェース         |
| `netbox_device_primary_ip`               | デバイスのプライマリ IP 設定     |
| `netbox_ip_address`                      | IP アドレス                      |
| `netbox_ip_address_range`                | IP アドレス範囲                  |
| `netbox_location`                        | ロケーション                     |
| `netbox_prefix`                          | プレフィックス                   |
| `netbox_site`                            | サイト                           |
| `netbox_tag`                             | タグ                             |
| `netbox_virtual_machine`                 | 仮想マシン                       |
| `netbox_virtual_machine_interface`       | 仮想マシンインターフェース       |
| `netbox_virtual_machine_primary_ip`      | 仮想マシンのプライマリ IP 設定   |
| `netbox_vlan`                            | VLAN                             |
| `netbox_vlan_group`                      | VLAN グループ                    |

### Data Sources

| データソース名                           | 説明                             |
| ---------------------------------------- | -------------------------------- |
| `netbox_custom_field`                    | カスタムフィールド (単体)        |
| `netbox_custom_fields`                   | カスタムフィールド (一覧)        |
| `netbox_device`                          | デバイス (単体)                  |
| `netbox_devices`                         | デバイス (一覧)                  |
| `netbox_ip_address`                      | IP アドレス (単体)               |
| `netbox_ip_addresses`                    | IP アドレス (一覧)               |
| `netbox_location`                        | ロケーション (単体)              |
| `netbox_locations`                       | ロケーション (一覧)              |
| `netbox_prefix`                          | プレフィックス (単体)            |
| `netbox_prefixes`                        | プレフィックス (一覧)            |
| `netbox_region`                          | リージョン (単体)                |
| `netbox_regions`                         | リージョン (一覧)                |
| `netbox_site`                            | サイト (単体)                    |
| `netbox_sites`                           | サイト (一覧)                    |
| `netbox_tag`                             | タグ (単体)                      |
| `netbox_tags`                            | タグ (一覧)                      |
| `netbox_virtual_machine`                 | 仮想マシン (単体)                |
| `netbox_virtual_machines`               | 仮想マシン (一覧)                |
| `netbox_vlan`                            | VLAN (単体)                      |
| `netbox_vlan_group`                      | VLAN グループ (単体)             |
| `netbox_vlan_groups`                     | VLAN グループ (一覧)             |
| `netbox_vlans`                           | VLAN (一覧)                      |

## Provider のビルド

```shell
git clone https://github.com/negineri/terraform-provider-netbox.git
cd terraform-provider-netbox
go install
```

ビルドした provider をローカルで使用するには、`~/.terraformrc` に以下を追加します。

```hcl
provider_installation {
  dev_overrides {
    "negineri/netbox" = "/path/to/GOPATH/bin"
  }

  direct {}
}
```

## 開発

### 依存関係の追加

```shell
go get github.com/author/dependency
go mod tidy
```

### ドキュメントの生成

```shell
make generate
```

### Acceptance テストの実行

Acceptance テストには稼働中の NetBox インスタンスが必要です。

```shell
make testacc
```

## ライセンス

[Mozilla Public License 2.0](LICENSE)
