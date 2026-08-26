# 文档检索索引服务

这是一个只使用 Go 标准库实现的后端检索 API 服务。它负责文档管理、分词、倒排索引、相关度排序、同义词扩展、拼写纠错、搜索建议、分类聚合和查询热度记录，不包含网页界面。

## 目录结构

- `cmd/server`：HTTP 服务入口和优雅关闭。
- `internal/handler`：API 路由、中间件和响应映射。
- `internal/service`：索引、搜索、重建、纠错和聚合逻辑。
- `internal/store`：带并发保护的内存数据访问层。
- `internal/model`：领域对象、状态和校验规则。
- `pkg`：日志、ID 和 HTTP 响应工具。

## 运行

```bash
go run ./cmd/server
```

服务默认监听 `:8080`。除健康检查外，API 请求需要携带 `X-Auth-Token`，默认值为 `dev-token`。

常用环境变量：

- `ADDR`：监听地址。
- `PORT`：未设置 `ADDR` 时使用的端口。
- `AUTH_TOKEN`：API 访问令牌。
- `MAX_PAGE_SIZE`：分页大小上限。
- `RATE_LIMIT`：单个来源每分钟请求上限。
- `LOG_LEVEL`：日志等级。

## 验证

```bash
go build ./...
go test ./...
```

## 主要接口

- `/api/documents`：文档写入、更新、读取和删除。
- `/api/indexes`：索引创建、状态切换和重建。
- `/api/index-documents`：把活动文档写入指定索引。
- `/api/search`：执行全文检索和相关度排序。
- `/api/search/facets`：按文档分类聚合搜索结果。
- `/api/suggest`：按前缀返回查询建议。
- `/api/correct`：返回拼写修正结果。
- `/api/synonyms`：维护查询扩展词。
- `/api/export/summary`：导出当前内存数据快照。

