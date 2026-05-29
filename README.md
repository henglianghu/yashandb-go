# yasdb-go|崖山数据库GO语言驱动程序

yasdb-go 是 YashanDB 官方 Go 语言数据库驱动，完全兼容 Go 标准的 `database/sql` 接口，依赖崖山数据库C驱动。

### 特性

- ✅ 全面兼容 Go 1.18+
- ✅ 符合 Go 标准 `database/sql` 接口
- ✅ 预处理语句与批量操作
- ✅ 事务支持
- ✅ 上下文控制（Cancel/Timeout）
- ✅ 多版本运行时检测与自动适配

### 环境准备（CGO + C 客户端）

本驱动依赖 YashanDB **C 客户端**，不同操作系统配置方式不同：

| 平台 | 在线安装 | 离线安装 |
|------|----------|----------|
| Linux | [docs/linux_online.md](docs/linux_online.md) | [docs/linux_offline.md](docs/linux_offline.md) |
| Windows | [docs/windows_online.md](docs/windows_online.md) | [docs/windows_offline.md](docs/windows_offline.md) |

- C 客户端下载：[yashandb-client](https://github.com/yashan-technologies/yashandb-client/releases)  
- Windows 另需 **64 位 GCC**（MinGW-w64 / TDM-GCC）以编译 CGO。

### 快速引入

```bash
go get github.com/yashan-technologies/yashandb-go@v1.4.2
```

```go
import _ "github.com/yashan-technologies/yashandb-go"

db, err := sql.Open("yasdb", "sys/password@127.0.0.1:1688")
```

更多示例见 [_examples](./_examples) 目录。

### DSN的填写说明

参考教程: [DSN format](./DSNFormat.md)

### 兼容性说明

| GO驱动版本 | 版本发布时间 | 新特性     | 最低兼容C驱动版本 | 完全支持C驱动版本 |
| ---------- | ------------ | ---------- | ----------------- | ----------------- |
| 1.4.2      | 2026.3.18    | 首版本支持 | v23.4.1.100       | v23.4.1.100       |

### 测试说明（向量 / 版本要求）

运行全量测试示例：

```bash
go test -v ./... -dsn "sys/password@127.0.0.1:1688"
```

- **向量（Vector）相关用例**（`TestVector*`）对版本有额外要求：需要 **YashanDB 23.5+**，并且需要配套的 **C 客户端（yashandb-client 23.5+）**。否则可能出现以下错误并导致全量测试失败：
  - **数据库不支持向量语法**：`YAS-04225 invalid word 3`
  - **C 客户端过旧**：`YAS-20001 symbol yacDescAlloc2 not found in yacli library`
- **全量测试耗时**：`TestManyQueryRowGosql` 会执行大量 `QueryRow`（默认 10000 次），在 Linux/远程网络环境下耗时数分钟属正常现象，可使用 `-short` 跳过该压力用例。

更完整的测试说明见：[docs/testing.md](docs/testing.md)。
