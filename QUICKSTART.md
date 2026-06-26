# Pritunl CLI - 快速开始指南

## 🚀 5 分钟快速上手

### 1. 初始化配置

```bash
# 交互式配置
./pritunl config init

# 或使用环境变量
export PRITUNL_BASE_URL=https://pritunl.example.com
export PRITUNL_API_TOKEN=your_token
export PRITUNL_API_SECRET=your_secret
```

### 2. 验证连接

```bash
./pritunl status
```

输出示例：
```
Key           Value
Server Version 1.32.4350.90
Organizations 5
Users         42
Users Online  12
...
```

### 3. 查看 Server

```bash
./pritunl server list
```

获取 Server ID（后续批量操作需要用到）。

### 4. 批量添加路由

#### 方式 A：使用 JSON

```bash
# 使用示例文件
./pritunl routes batch-add <server-id> -file examples/routes_sample.json

# 或自己的文件
./pritunl routes batch-add <server-id> -file my_routes.json
```

#### 方式 B：使用 CSV

```bash
./pritunl routes batch-add <server-id> -csv examples/routes_sample.csv
```

#### 方式 C：跳过确认

```bash
./pritunl routes batch-add <server-id> -file routes.json -skip-confirm
```

### 5. 验证结果

```bash
# 查看已添加的路由
./pritunl routes list <server-id>

# JSON 格式
./pritunl routes list <server-id> -output json
```

## 📋 常见命令

### Status

```bash
./pritunl status              # 查看服务器状态
./pritunl status -output json # JSON 格式
```

### Server 管理

```bash
./pritunl server list         # 列表
./pritunl server get <id>     # 详情
./pritunl server start <id>   # 启动
./pritunl server stop <id>    # 停止
```

### Routes 管理

```bash
# 查看路由
./pritunl routes list <id>

# 单条添加
./pritunl routes add <id> -network 10.0.0.0/24 -comment "Office"

# 批量添加
./pritunl routes batch-add <id> -file routes.json

# 验证文件
./pritunl routes validate -file routes.json

# 导出/备份
./pritunl routes export <id> -file backup.json

# 删除路由
./pritunl routes delete <id> -network 10.0.0.0/24
```

### Organization 和 User

```bash
./pritunl org list              # 组织列表
./pritunl org get <id>          # 组织详情
./pritunl user list <org-id>    # 用户列表
./pritunl user get <org> <user> # 用户详情
```

### 配置管理

```bash
./pritunl config init           # 交互式初始化
./pritunl config show           # 显示配置
./pritunl config set key value  # 设置单个值
./pritunl config clear          # 清除配置
```

## 📝 数据格式

### JSON 路由文件

```json
[
  {
    "network": "10.0.0.0/24",
    "comment": "Office Network",
    "metric": 100,
    "nat": false
  },
  {
    "network": "192.168.1.0/24",
    "comment": "Remote Site",
    "nat": true,
    "nat_interface": "eth0"
  }
]
```

### CSV 路由文件

```
network,comment,metric,nat,nat_interface
10.0.0.0/24,Office Network,100,false,
192.168.1.0/24,Remote Site,100,true,eth0
```

## 🔄 完整工作流：从 Excel 导入路由

### 1. 在 Excel 中准备数据

创建包含以下列的 Excel：
- network (CIDR 格式)
- comment (描述)
- metric (优先级)
- nat (true/false)
- nat_interface (网卡名)

### 2. 导出为 CSV

```
network,comment,metric,nat,nat_interface
10.0.0.0/24,Office,100,false,
10.0.1.0/24,Lab,100,false,
192.168.1.0/24,Remote,100,true,eth0
```

### 3. 验证格式

```bash
./pritunl routes validate -csv my_routes.csv
```

### 4. 批量导入

```bash
# 先查看预览
./pritunl routes batch-add <server-id> -csv my_routes.csv

# 确认后自动添加，或跳过确认
./pritunl routes batch-add <server-id> -csv my_routes.csv -skip-confirm
```

### 5. 验证结果

```bash
./pritunl routes list <server-id>
```

## 💡 使用技巧

### 1. 保存配置到文件

```bash
./pritunl config init
# 配置保存在 ~/.pritunl/config.yaml
cat ~/.pritunl/config.yaml
```

### 2. 使用环境变量 (临时)

```bash
export PRITUNL_BASE_URL=https://pritunl.example.com
export PRITUNL_API_TOKEN=xxx
export PRITUNL_API_SECRET=yyy
./pritunl status
```

### 3. 批量操作多个 Server

```bash
# 导出服务器 A 的路由
./pritunl routes export server-a-id -file backup.json

# 导入到服务器 B
./pritunl routes batch-add server-b-id -file backup.json -skip-confirm
```

### 4. 输出格式对比

```bash
# 默认表格格式
./pritunl server list

# JSON 格式（方便脚本处理）
./pritunl server list -output json

# YAML 格式
./pritunl server list -output yaml
```

### 5. 一行命令获取 Server ID

```bash
# 获取第一个 Server 的 ID (适用于 bash/jq)
SERVER_ID=$(./pritunl server list -output json | head -1 | jq -r '.id')
echo $SERVER_ID
```

## ⚠️ 常见问题

### Q: 出现 "Server online" 错误

A: 需要先停止 Server：
```bash
./pritunl server stop <server-id>
./pritunl routes batch-add <server-id> -file routes.json
./pritunl server start <server-id>
```

### Q: 如何跳过确认提示？

A: 使用 `-skip-confirm` 标志：
```bash
./pritunl routes batch-add <id> -file routes.json -skip-confirm
```

### Q: 配置文件在哪？

A: `~/.pritunl/config.yaml`

### Q: 如何修改配置？

A:
```bash
# 修改单个值
./pritunl config set base_url https://new-url.com

# 或编辑文件
nano ~/.pritunl/config.yaml
```

### Q: 支持多少条路由一次添加？

A: 理论无限制。建议 100-500 条为一批。

## 🔍 调试

### 验证 API 连接

```bash
./pritunl status
```

### 验证路由文件

```bash
./pritunl routes validate -file routes.json
```

### 查看详细错误

```bash
./pritunl routes batch-add <id> -file routes.json 2>&1 | head -50
```

### 检查配置

```bash
./pritunl config show
```

## 📚 更多信息

详见 `README.md` 获取完整文档。

### 关键文件

- `pritunl` - CLI 可执行文件
- `examples/config.yaml` - 配置文件示例
- `examples/routes_sample.json` - JSON 示例
- `examples/routes_sample.csv` - CSV 示例
- `README.md` - 完整文档

---

**现在就开始使用吧！** 🎉

```bash
# 3 个简单步骤

# 1. 配置
./pritunl config init

# 2. 验证
./pritunl status

# 3. 批量添加路由
./pritunl routes batch-add <server-id> -file examples/routes_sample.json
```
