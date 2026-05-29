# Windows 在线安装说明

本驱动通过 **CGO** 调用 YashanDB **C 客户端**。在 Windows 上开发或运行前，需安装 **64 位 GCC（MinGW）** 并配置 **C 客户端 `lib` 目录到 PATH**。

> **模块路径**：`github.com/yashan-technologies/yashandb-go`（与 `go.mod` 一致）。  
> **C 驱动版本**：参见主仓 [README](../README.md) 兼容性表（当前 Go 驱动 1.4.2 最低兼容 C 驱动 v23.4.1.100）。

---

## 1. 安装 64 位 GCC（CGO 编译需要）

可从以下任选其一安装 **64 位** 工具链：

- [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download)
- [MinGW-w64](https://www.mingw-w64.org/downloads/)

以 TDM-GCC 为例：

1. 运行安装程序，选择安装目录（如 `C:\TDM-GCC-64`）。
2. 确认安装目录下的 `bin`（如 `C:\TDM-GCC-64\bin`）已加入系统 **PATH**；若使用其他 MinGW 发行版，请手动添加。
3. 重新打开 **CMD** 或 **PowerShell**，执行：

```cmd
gcc --version
```

无报错即表示可用。

---

## 2. 安装 YashanDB C 客户端（Windows）

1. 打开 [yashandb-client](https://github.com/yashan-technologies/yashandb-client/releases)，下载与 README 兼容表匹配的 **`yashandb-client-<版本>-win-x86_64.zip`**（或 `.tar.gz`，按发布页实际文件名）。
2. 解压到本地目录，例如：`D:\yashandb\yashandb-client`。
3. 将解压目录下的 **`lib`** 文件夹路径加入系统环境变量 **PATH**，例如：

```text
D:\yashandb\yashandb-client\lib
```

4. 修改环境变量后需**重新打开**终端，再执行 `go build` / `go run`。

> 也可从 YashanDB 产品安装包的 `Drivers` 目录获取同版本 Windows C 客户端包，步骤相同。

---

## 3. 创建示例项目

```cmd
mkdir yashandb_connect
cd yashandb_connect
```

创建 `main.go`：

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
	err = db.QueryRow("SELECT STATUS FROM V$DATABASE").Scan(&status)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("STATUS:", status)
}
```

---

## 4. 编译与运行

```cmd
go mod init yashandb_connect
go get github.com/yashan-technologies/yashandb-go@latest
go mod tidy
go run main.go
```

指定版本（推荐生产环境）：

```cmd
go get github.com/yashan-technologies/yashandb-go@v1.4.2
go mod tidy
```

---

## 5. 运行本驱动源码（开发/测试）

在克隆的 `yashandb-go` 目录中执行测试前，同样需完成 **GCC** 与 **C 客户端 PATH** 配置：

```cmd
cd path\to\yashandb-go
go test ./... -dsn="sys/password@127.0.0.1:1688"
```

若出现 `undefined: YasConn` 等 CGO 相关错误，多为 **GCC 未安装/未在 PATH**，或 **C 客户端 `lib` 未加入 PATH**。

---

## 6. 分发可执行文件注意事项

将依赖本驱动的程序拷贝到其他 Windows 机器运行时，除可执行文件外，通常还需一并提供：

- YashanDB C 客户端 **`lib` 目录下的 DLL**（具体文件名以 C 客户端包内为准，常见为 `yascli`、`yas_infra` 等相关库）；
- 若 C 客户端依赖 OpenSSL 等第三方 DLL，也需按 C 客户端说明一并部署。

可将上述 DLL 与可执行文件放在同一目录，或确保目标机器的 **PATH** 包含这些 DLL 所在目录。

---

## 7. 无法访问 GitHub 时（内网镜像）

若环境无法访问 `github.com`，可在应用侧保持 `module` 为 `github.com/yashan-technologies/yashandb-go`，通过 Git `insteadOf` 将克隆地址指向公司内网镜像仓。具体见团队内部《双仓协作与迁移使用指南》，**无需**修改 `go.mod` 的 module 行。
