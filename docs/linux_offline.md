# Linux 离线安装说明

适用于无法通过 `go get` 从 GitHub 拉取模块、但已在本地获得 **yashandb-go 源码包** 的场景。

---

## 1. 获取源码

- 从 [Releases](https://github.com/yashan-technologies/yashandb-go/releases) 下载源码归档；或  
- 使用 `git clone` 后离线拷贝；或  
- 使用公司内网 GitLab 镜像仓同步后的源码。

解压到例如：`/opt/src/yashandb-go`。

---

## 2. 配置 C 客户端

与 [linux_online.md](./linux_online.md) 相同：安装 **linux-x86_64** C 客户端并设置：

```bash
export LD_LIBRARY_PATH=/opt/yashandb/yashandb-client/lib:$LD_LIBRARY_PATH
```

---

## 3. 使用 go mod replace

```bash
mkdir yashandb_connect && cd yashandb_connect
# 编写 main.go（import 为 github.com/yashan-technologies/yashandb-go）

go mod init yashandb_connect
go mod edit -replace github.com/yashan-technologies/yashandb-go=/opt/src/yashandb-go
go mod tidy
go run main.go
```

---

## 4. 在本仓库目录测试

```bash
cd /opt/src/yashandb-go
go test ./... -dsn="sys/password@127.0.0.1:1688"
```

---

## 5. 分发说明

部署时需包含 C 客户端 `lib` 下的 `.so` 依赖，参见 [linux_online.md](./linux_online.md) 第 5 节。
