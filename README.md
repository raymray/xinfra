# xinfra

Go 微服务基础库，配合 xframe 服务模板使用。各包由调用方构造并持有，不提供包级全局句柄。

## 包

| 包 | 内容 |
|---|---|
| `xerror` | 带错误码与堆栈的错误包装，兼容 `errors.Is` / `errors.As` |
| `xlog` | zap 封装、文件轮转（`xlog/xrotate`）、ctx 内 logger |
| `xdb` | GORM 多实例管理，预连接与慢 SQL 日志 |
| `xredis` | `redis.UniversalClient` 多实例管理，支持 single / sentinel / cluster |
| `xlifecycle` | 分阶段关闭与退出信号处理 |
| `xhttp` | 统一响应信封、错误码映射、request id |
| `xhttp/xgin` | gin 适配：`OK` / `Fail`、Recovery、Trace、AccessLog |
| `xkit` | 通用工具：`BiMap`、`MChannel`、时间解析；根包无第三方依赖 |
| `xkit/fskit` | CSV、Excel、JSON 读写 |
| `xkit/xresty` | resty 请求封装 |

## 使用

```bash
go get github.com/raymray/xinfra
```

本地联调时，在服务的 `go.mod` 中指向本地目录：

```text
replace github.com/raymray/xinfra => ../xinfra
```

## 测试

```bash
go vet ./... && go test ./...
```
