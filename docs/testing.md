# 测试指南（yashandb-go）

本文档说明如何在本地/CI 环境运行本项目测试，以及常见的“版本不匹配”导致的失败原因与处理方式。

> 注意：本驱动依赖 CGO 与 YashanDB C 客户端（`yashandb-client`）。请先按 `README.md` 和 `docs/*_online.md` 完成环境准备。

---

## 1. 最小验证（推荐）

只验证能连库与基本驱动可用：

```bash
go test -v -run TestConnect . -dsn "user/password@host:port"
```

---

## 2. 运行全量测试

```bash
go test -v ./... -dsn "user/password@host:port"
```

若你使用 PowerShell，参数也可以写成：

```powershell
go test -v ./... "-dsn=user/password@host:port"
```

---

## 3. 向量（Vector）测试的版本要求

向量相关用例位于 `sql_vector_test.go`，用例名以 `TestVector` 开头。

### 3.1 数据库版本要求

若目标数据库版本 **不支持向量类型语法**，`CREATE TABLE ... vector(...)` 会报错并导致用例失败，常见错误：

- `YAS-04225 invalid word 3`

解决：升级数据库到支持向量能力的版本（建议 **YashanDB 23.5+**）。

### 3.2 C 客户端版本要求

即使数据库支持向量，如果 **C 客户端版本过旧**，驱动在运行时动态加载符号失败，会报错：

- `YAS-20001 symbol yacDescAlloc2 not found in yacli library`

解决：升级 `yashandb-client` 至与数据库版本线匹配的版本（建议 **yashandb-client 23.5+**）。

---

## 4. 其它按能力跳过（Skip）的用例

- `TIMESTAMP WITH TIME ZONE`：低版本数据库可能不支持，相关用例会 `t.Skip`（不计入失败）。
- `Uint`（无符号整型）：需要 MySQL 兼容模式；若处于 YashanDB 模式，相关用例会 `t.Skip`。

---

## 5. “看起来卡住不动”的常见原因

### 5.1 `TestManyQueryRowGosql` 耗时较长

该用例默认执行 10000 次 `QueryRow`，且循环中几乎不输出日志，因此在 Linux/远程网络环境下可能表现为“很久没新输出”，但实际上仍在运行。

可选处理：

- 加超时：`go test -timeout 10m ...`
- 关闭该压力用例：`go test -short ...`

### 5.2 并行用例集中运行

本项目大量测试使用 `t.Parallel()`。Go 测试框架会先收集并行用例，等串行用例完成后再集中并发执行，也可能出现一段时间输出很少。

可选处理：

- 降低并行度：`go test -parallel 1 ...`

---

## 6. 建议的测试执行顺序

1. `TestConnect`（确认 DSN、网络、权限、C 客户端环境）
2. 关键路径小集合（事务、DDL/DML、PL/SQL debug）
3. 全量 `./...`

