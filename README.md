# Pritunl CLI

完整功能的 Pritunl VPN 管理工具，基于 Go SDK 构建。支持 **批量路由操作**、配置管理、多种输出格式等功能。

## ✨ 特性

- 🚀 **批量路由操作**：从 JSON/CSV 文件一次添加多条路由
- 🔧 **完整的资源管理**：Server、Organization、User、Routes
- 📊 **多种输出格式**：Table、JSON、YAML
- ⚙️ **配置文件支持**：`~/.pritunl/config.yaml`
- ✅ **验证和冲突检测**：自动检测重复路由和格式错误
- 💾 **导入/导出**：备份和恢复路由配置
- 🔐 **安全**：使用 HMAC-SHA256 认证（基于 SDK）

## 📦 安装

### 从源代码构建

```bash
# 克隆仓库
git clone https://github.com/zhong/pritunl-cli
cd pritunl-cli

# 构建
go build -o pritunl main.go

# 移到 PATH
sudo mv pritunl /usr/local/bin/
```

### 快速验证

```bash
pritunl version
pritunl help
```

## 🚀 快速开始

### 1. 初始化配置

```bash
pritunl config init
```

按照提示输入：
- Pritunl 服务器 URL
- API Token
- API Secret

也可以使用环境变量：

```bash
export PRITUNL_BASE_URL=https://pritunl.example.com
export PRITUNL_API_TOKEN=your_token
export PRITUNL_API_SECRET=your_secret
```

### 2. 验证连接

```bash
pritunl status
```

### 3. 批量添加路由

```bash
# 从 JSON 文件
pritunl routes batch-add <server-id> -file routes.json

# 从 CSV 文件
pritunl routes batch-add <server-id> -csv routes.csv

# 跳过确认提示
pritunl routes batch-add <server-id> -file routes.json -skip-confirm
```

## 📖 完整命令参考

### Status

```bash
pritunl status                          # 获取服务器状态
pritunl status -output json             # JSON 格式
```

### Server 管理

```bash
pritunl server list                     # 列表
pritunl server list -output json        # JSON 格式
pritunl server get <id>                 # 获取详情
pritunl server start <id>               # 启动
pritunl server stop <id>                # 停止
pritunl server restart <id>             # 重启
```

### Organization 管理

```bash
pritunl org list                        # 列表
pritunl org get <id>                    # 获取详情
```

### User 管理

```bash
pritunl user list <org-id>              # 列表
pritunl user get <org-id> <user-id>     # 获取详情
```

### Routes 管理（核心功能）

```bash
# 列表
pritunl routes list <server-id>
pritunl routes list <server-id> -output json

# 单条添加
pritunl routes add <server-id> -network 10.0.0.0/24 -comment "Office"

# 批量添加
pritunl routes batch-add <server-id> -file routes.json
pritunl routes batch-add <server-id> -csv routes.csv
pritunl routes batch-add <server-id> -file routes.json -skip-confirm

# 验证文件
pritunl routes validate -file routes.json
pritunl routes validate -csv routes.csv

# 导出/备份
pritunl routes export <server-id> -file backup.json
pritunl routes export <server-id> -file backup.csv

# 删除
pritunl routes delete <server-id> -network 10.0.0.0/24
```

### 配置管理

```bash
pritunl config init                     # 交互式初始化
pritunl config show                     # 显示当前配置
pritunl config set base_url https://...  # 设置单个值
pritunl config clear                    # 清除配置
```

## 📋 JSON 路由格式

```json
[
  {
    "network": "10.0.0.0/24",
    "comment": "Office Network",
    "metric": 100,
    "nat": false,
    "nat_interface": "",
    "advertise": false,
    "net_gateway": false
  },
  {
    "network": "192.168.1.0/24",
    "comment": "Remote Site",
    "metric": 100,
    "nat": true,
    "nat_interface": "eth0"
  }
]
```

## 📊 CSV 路由格式

```
network,comment,metric,nat,nat_interface,advertise,net_gateway
10.0.0.0/24,Office Network,100,false,,false,false
192.168.1.0/24,Remote Site,100,true,eth0,false,false
```

## 🔄 工作流示例

### 场景 1：批量添加路由

```bash
# 1. 准备路由文件 (routes.json 或 routes.csv)
# 2. 验证格式
pritunl routes validate -file routes.json

# 3. 预览将要添加的路由
pritunl routes batch-add <server-id> -file routes.json
# 系统会显示预览并要求确认

# 4. 批量添加
pritunl routes batch-add <server-id> -file routes.json -skip-confirm

# 5. 验证结果
pritunl routes list <server-id>
```

### 场景 2：迁移路由配置

```bash
# 1. 从旧 Server 导出路由
pritunl routes export <old-server-id> -file backup.json

# 2. 编辑或保持原样
# nano backup.json

# 3. 导入到新 Server
pritunl routes batch-add <new-server-id> -file backup.json
```

### 场景 3：从 Excel 导入

```bash
# 1. Excel 中准备数据，导出为 CSV
# 2. 验证格式
pritunl routes validate -csv routes.csv

# 3. 导入
pritunl routes batch-add <server-id> -csv routes.csv
```

## 🔧 环境变量

```bash
PRITUNL_BASE_URL        # Pritunl 服务器地址
PRITUNL_API_TOKEN       # API Token
PRITUNL_API_SECRET      # API Secret
PRITUNL_INSECURE        # 跳过 TLS 验证 (true/false)
PRITUNL_OUTPUT_FORMAT   # 输出格式 (table/json/yaml)
```

## 📝 配置文件

位置：`~/.pritunl/config.yaml`

```yaml
base_url: https://pritunl.example.com
api_token: your_token_32_chars
api_secret: your_secret_32_chars
insecure: true
output_format: table
default_server_id: 5f1234567890abcdef000000
```

## ⚠️ 常见问题

### Q: 如何启用 API Token？
A: 在 Pritunl 服务器上运行：
```bash
python3 enable_pritunl_api_token_nopymongo.py
```

### Q: 批量添加时出现 "Server online" 错误？
A: Server 必须处于离线状态才能修改路由。在 Web UI 中停止该 Server。

### Q: 如何验证路由文件格式？
A: 使用 validate 命令：
```bash
pritunl routes validate -file routes.json
```

### Q: 支持多少条路由一次添加？
A: 理论无限制，建议按 100-500 条分批。

### Q: 可以自动跳过确认提示吗？
A: 是的，使用 `-skip-confirm` 标志：
```bash
pritunl routes batch-add <id> -file routes.json -skip-confirm
```

## 📊 输出格式

### Table（默认）
```
ID                      Name                  Status  Network
5f1234567890abcdef000000 VPN-Server-01 online 10.0.0.0/24
```

### JSON
```json
{
  "id": "5f1234567890abcdef000000",
  "name": "VPN-Server-01",
  "status": "online"
}
```

### YAML
```yaml
- id: 5f1234567890abcdef000000
  name: VPN-Server-01
  status: online
```

## 🏗️ 项目结构

```
pritunl-cli/
├── main.go                      # 入口点
├── go.mod                       # 依赖
├── cmd/                         # 命令实现
│   ├── root.go
│   ├── status.go
│   ├── server.go
│   ├── org.go
│   ├── user.go
│   ├── routes.go                # 批量路由操作核心
│   ├── config.go
│   ├── util.go
│   └── errors.go
├── pkg/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── output/
│   │   └── formatter.go         # 输出格式控制
│   ├── routes/
│   │   ├── loader.go            # 路由加载（JSON/CSV）
│   │   ├── validator.go         # 验证逻辑
│   │   └── batch.go             # 批量操作
│   └── util/
├── examples/                    # 示例配置和数据
│   ├── config.yaml
│   ├── routes_sample.json
│   └── routes_sample.csv
└── README.md
```

## 🔐 API 认证

CLI 使用与 Go SDK 相同的 HMAC-SHA256 认证机制：

```
签名原文: API_TOKEN & TIMESTAMP & NONCE & METHOD & PATH
签名值: base64(HMAC-SHA256(API_SECRET, auth_string))
```

详见 SDK 文档。

## 📚 相关文档

- [Pritunl Go SDK](../pritunl-go-sdk/README.md)
- [Python 批量路由工具](../batch_routes_solution.md)
- [Pritunl 官方文档](https://docs.pritunl.com)

## 🐛 故障排除

### 连接失败

```bash
# 检查 URL 和凭证
pritunl config show

# 测试连接
pritunl status

# 检查环境变量
echo $PRITUNL_BASE_URL $PRITUNL_API_TOKEN
```

### 路由添加失败

```bash
# 验证路由文件格式
pritunl routes validate -file routes.json

# 查看错误信息
pritunl routes batch-add <id> -file routes.json 2>&1 | head -20
```

### 输出格式问题

```bash
# 指定输出格式
pritunl server list -output json
pritunl server list -output yaml
```

## 💡 最佳实践

1. **备份**：修改前总是导出当前配置
   ```bash
   pritunl routes export <id> -file backup.json
   ```

2. **验证**：批量操作前验证文件
   ```bash
   pritunl routes validate -file routes.json
   ```

3. **分批**：大量路由分批添加（100-500 条为宜）

4. **确认**：关键操作前检查预览

5. **监控**：操作后检查结果
   ```bash
   pritunl routes list <id>
   ```

## 📄 许可证

本项目基于 Pritunl SDK，遵循其许可证条款。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 支持

- 问题报告：GitHub Issues
- 文档：本 README 和示例代码
- 相关项目：Pritunl Go SDK

---

**版本**：1.0.0  
**最后更新**：2026-06-26
