# Linux 在线安装说明

本驱动通过 **CGO** 调用 YashanDB **C 客户端**。在 Linux 上开发或运行前，需将 C 客户端的 **`lib` 目录** 加入动态库搜索路径（`LD_LIBRARY_PATH`）。

> **模块路径**：`github.com/yashan-technologies/yashandb-go`  
> **C 驱动版本**：参见主仓 [README](../README.md) 兼容性表。

---

## 1. 安装 YashanDB C 客户端（Linux）

1. 打开 [yashandb-client](https://github.com/yashan-technologies/yashandb-client/releases)，下载 **`yashandb-client-<版本>-linux-x86_64.tar.gz`**（版本与 README 兼容表一致）。
2. 解压，例如：

```bash
tar -xzf yashandb-client-23.4.7.100-linux-x86_64.tar.gz -C /opt/yashandb
```

3. 设置动态库路径（按实际解压路径修改）：

```bash
export LD_LIBRARY_PATH=/opt/yashandb/yashandb-client/lib:$LD_LIBRARY_PATH
```

常见依赖库（由 C 客户端包提供，名称以包内为准）包括：

- `libyas_infra.so.0`
- `libyascli.so.0`
- 以及 `libcrypto`、`libpcre` 等（若包内 `lib` 已包含则无需单独安装）

建议将 `export LD_LIBRARY_PATH=...` 写入 `~/.bashrc` 以便长期生效。

---

## 2. 创建示例项目

```bash
mkdir yashandb_connect && cd yashandb_connect
```

`main.go`：

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/yashan-technologies/yashandb-go"
)

func main() {
	db, err := sql.Open("yasdb", "sys/password@127.0.0.1:1688")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var status string
	if err := db.QueryRow("SELECT STATUS FROM V$DATABASE").Scan(&status); err != nil {
		log.Fatal(err)
	}
	log.Println("STATUS:", status)
}
```

---

## 3. 编译与运行

```bash
go mod init yashandb_connect
go get github.com/yashan-technologies/yashandb-go@v1.4.2
go mod tidy
go run main.go
```

---

## 4. 运行本驱动源码测试

```bash
cd /path/to/yashandb-go
go test ./... -dsn="sys/password@127.0.0.1:1688"
```

---

## 5. 分发可执行文件

将程序部署到其他 Linux 主机时，需一并提供或在该机安装相同版本的 C 客户端，并确保运行时能找到：

- `libyas_infra.so.0`
- `libyascli.so.0`

等（以 C 客户端 `lib` 目录实际文件为准）。可通过 `LD_LIBRARY_PATH` 或 `rpath` 指定库路径。

---

## 6. 无法访问 GitHub 时（内网镜像）

保持 `go.mod` 中 `module` / `require` 为 `github.com/yashan-technologies/yashandb-go`，通过 `git config url.insteadOf` 将克隆地址映射到内网 GitLab 镜像。详见团队《双仓协作与迁移使用指南》。
