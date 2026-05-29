# Windows 离线安装说明

适用于无法通过 `go get` 从 GitHub 拉取模块、但已在本地获得 **yashandb-go 源码包** 的场景。CGO 与 C 客户端要求与 [在线说明](./windows_online.md) 相同。

---

## 1. 获取源码

任选其一：

- 从 [yashandb-go Releases](https://github.com/yashan-technologies/yashandb-go/releases) 下载源码归档；或
- `git clone` 后打包拷贝到离线机器；或
- 使用公司内网 GitLab 镜像仓同步后的源码（`module` 仍为 `github.com/yashan-technologies/yashandb-go`）。

将源码解压到本地路径，例如：`D:\src\yashandb-go`。

---

## 2. 环境准备（与在线说明相同）

1. 安装 **64 位 GCC**（TDM-GCC 或 MinGW-w64），并确保 `gcc` 在 **PATH** 中。  
2. 安装 **YashanDB C 客户端（win-x86_64）**，将 **`lib` 目录** 加入 **PATH**。  

详见 [windows_online.md](./windows_online.md) 第 1、2 节。

---

## 3. 创建应用项目

```cmd
mkdir yashandb_connect
cd yashandb_connect
```

`main.go` 示例（import 路径与线上一致）：

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
	log.Println("connected")
}
```

---

## 4. 使用 go mod replace 指向本地源码

```cmd
go mod init yashandb_connect

go mod edit -replace github.com/yashan-technologies/yashandb-go=D:\src\yashandb-go
go mod tidy

go run main.go
```

将 `D:\src\yashandb-go` 替换为你本机解压后的**绝对路径**。

也可在 `go.mod` 中手写：

```go
replace github.com/yashan-technologies/yashandb-go => D:/src/yashandb-go
```

---

## 5. 直接在本仓库目录开发/测试

```cmd
cd D:\src\yashandb-go
go test ./... -dsn="sys/password@127.0.0.1:1688"
```

---

## 6. 分发可执行文件

与 [windows_online.md](./windows_online.md) 第 6 节相同：部署时需包含 C 客户端 **`lib` 下的 DLL** 及必要依赖库。
