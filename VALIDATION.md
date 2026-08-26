# gb-502 验收记录

验收日期：2026-08-22（Asia/Shanghai）

## 结论

通过。项目符合原始提示词的水产养殖领域、分层结构、共享组件、权限体系、业务约束和部署要求，可通过 Docker Compose 正常启动。四实体业务链、五个核心页面、管理员主流程和观察员 RBAC 均已真实验证，没有发现需要修改业务代码的问题。

## 原始要求逐项核验

| 原始要求 | 实际核验 | 结果 |
| --- | --- | --- |
| 水质与投喂控制领域，排除电商/库存/记账/预约 | README、页面、API 和实体只覆盖养殖池、水质、计划、投喂执行与审计，未发现越界模块 | 通过 |
| 四个核心实体全链路贯穿 | `Pond`、`WaterReading`、`FeedingPlan`、`ControlExecution` 均有独立 model、DTO、repository、service、handler、API 和页面文件；README 有逐层映射表 | 通过 |
| 五个核心页面 | `/ponds`、`/readings`、`/plans`、`/executions`、`/audit` 均已在内置 Browser 中加载并检查真实数据 | 通过 |
| `RiskTag` 跨页共用 | `PondsPage.vue` 与 `ReadingsPage.vue` 均引用；Browser 中可见正常和预警状态 | 通过 |
| `PlanDrawer` 跨页共用 | `PlansPage.vue` 与 `ExecutionsPage.vue` 均引用；两个页面都实际打开了同一计划详情抽屉 | 通过 |
| JWT/RBAC 全链路 | 数据库角色、JWT claims、Go auth/RBAC 中间件、Vue 路由 `meta.roles`、导航显隐和操作显隐联动；`viewer` 无新增按钮和审计入口，直达 `/audit` 被重定向；后端写接口和审计接口均返回 403 | 通过 |
| 规则版本审计 | 计划创建、草稿修订自动升版、提交、批准、撤销与执行均在 service 中记录前后状态及原因；真实审计数据可见 create/revise/submit/approve/execute 链路 | 通过 |
| 全局错误、请求 ID、Redis 限流独立实现 | recovery/error 响应、请求 ID 和 Redis rate limit 均有独立中间件；响应含 `X-Request-ID` 和限流头 | 通过 |
| 共享枚举一致 | Go 与 TypeScript 中 `PondStatus` 为 `active/quarantine/closed`，`PlanStatus` 为 `draft/pending/approved/executed`；README 明确同步位置 | 通过 |
| 至少三个公共组件、两个 hooks | `StatusBadge`、`MetricCard`、`ConfirmDialog` 齐全且被页面引用；额外提供 `RiskTag`、`PlanDrawer`；`useAuth`、`useQueryParams` 均实际使用 | 通过 |
| Vue 3 + TypeScript + Vite + Element Plus | `package.json`、源码和生产构建结果一致 | 通过 |
| Go 1.22 + Gin + GORM + PostgreSQL + Redis | `go.mod`、后端实现、Compose 服务与健康检查一致 | 通过 |
| 2800-4000 行、28-40 个 Go 文件，至少 20 个功能文件 | 原提示词未要求排除 `_test.go`：仓库共 2826 行、40 个 `.go` 文件；排除测试后为 2701 行、39 个功能文件，明显超过至少 20 个功能文件要求 | 通过 |
| 强制目录与职责分层 | 前端八类目录和后端 `model/dto/repository/service/handler/router/middleware/constants/util` 均存在；实体与职责未合并为单文件 | 通过 |
| Compose、环境文件与反向代理 | 顶层 `name` 正确且无 `version`；`.env`/`.env.example` 均有项目名；端口为 18502/19502；Nginx 将 `/api` 转发至 `backend:8080` | 通过 |
| 数据库命名卷、健康检查、依赖等待 | PostgreSQL/Redis 使用命名卷和 healthcheck，后端等待二者 healthy，前端等待后端 healthy | 通过 |
| 真实 `/healthz`、README、Git | 后端与 Nginx 同源 `/healthz` 均返回 200 且数据库/Redis 为 `ok`；README 完整；本地分支最终为 `main`，保留原提交 `cf4bc76` | 通过 |

## 命令验证

以下命令在最终代码上执行并返回退出码 0：

```bash
cd backend
go test ./...
go vet ./...
go build ./...

cd ../frontend
npm run build

cd ..
docker compose config --quiet
docker compose up -d --build
docker compose ps
```

运行时 PostgreSQL、Redis、backend 均报告 `healthy`，frontend 正常运行在 `18502`。前端生产构建成功；Vite 仅有主包超过 500 kB 的性能建议，不阻断功能或部署。

规模统计结果：

```text
全部 Go 文件：40 个 / 2826 行
排除 _test.go：39 个 / 2701 行
```

## API 与业务红线

- `GET http://localhost:19502/healthz`：`200 OK`，`checks.database=ok`、`checks.redis=ok`。
- `GET http://localhost:18502/healthz`：经 Nginx 转发后同样为 `200 OK`。
- 使用 `viewer` JWT 请求 `GET /api/audit`：`403 Forbidden`，错误码 `FORBIDDEN`，响应带请求 ID。
- 使用 `viewer` JWT 请求 `POST /api/ponds`：`403 Forbidden`，未创建数据。
- 创建临时 `closed` 养殖池后请求 `POST /api/readings`：`409 Conflict`，消息为“已关闭养殖池不能新增读数”；临时池随后成功删除（204）。
- 对状态为 `executed` 的计划再次请求 `POST /api/executions`：`409 Conflict`，消息为“只能使用已批准计划安排执行”。
- 现有端到端记录展示了正常读数、异常读数确认、计划创建/提交/批准、执行安排/开始/完成、计划同步为 `executed` 以及完整审计链。

## 内置 Browser 验证

只使用 Codex 内置 Browser，未使用外部 Chrome。

- 登录页：管理员和观察员账号均可正常登录，路由跳转正确。
- `/ponds`：统计卡、`RiskTag`、状态和三条池塘数据正常；搜索 `P-SMOKE-502` 后只返回目标池；编辑弹窗字段回填正确，取消不写入。
- `/readings`：正常/预警指标、人工确认状态和两条读数正常；“录入读数”弹窗字段完整，取消不写入。
- `/plans`：计划版本、周期、状态正常；点击计划打开 `PlanDrawer`，可查看阈值、审核人与制定依据。
- `/executions`：已完成执行显示计划/实际量、天气和操作人；从执行记录再次打开 `PlanDrawer` 成功。
- `/audit`：17 条真实审计记录覆盖四类实体；详情抽屉展示操作人、请求 ID、原因和变更前后 JSON。
- `viewer` 登录后“新建养殖池”按钮数量为 0，审计导航数量为 0；直达 `/audit` 后实际 URL 为 `/ponds`。
- 整个最终 Browser 会话的 console error/warn 数量为 0。

## 清理状态

验收后已执行：

```bash
docker compose down -v --remove-orphans
```

`docker compose ps -a` 为空；按项目名过滤的容器和卷均为空。测试数据卷已按任务要求删除，不可从本次 Compose 环境恢复。

## 2026-08-22 企业级复核补充

- 养殖池、水质读数、计划和执行的全部写操作现以 PostgreSQL `SERIALIZABLE` 事务运行，业务变更和审计同提交、同回滚；序列化冲突与死锁最多自动重试 3 次。
- 计划、执行、读数和养殖池状态更新增加行锁，阻止重复批准、重复完成及丢失更新。真实并发提交同一执行反馈得到 `200,409`，仅一次业务完成成功。
- 执行状态机禁止 `running -> scheduled` 逆向迁移；`completed/cancelled` 保持终态。真实 API 逆向请求返回 409。
- 同一 UTC 日的未取消安排会累计核算，不得超过计划日投喂量；实测已有 80 kg 安排后再排 30 kg（计划日限 100 kg）返回 409。
- 周期计划仅在最后一条待执行/执行中记录完成后转为 `executed`；实测第一条完成后仍为 `approved`，第二条完成后才为 `executed`。
- 计划批准要求 24 小时内水质；`critical` 水质即使已确认留痕也不能批准。实测确认严重氨氮异常后批准仍返回 409。
- 登录增加独立 10 次/分钟限流且 Redis 故障时登录保护 fail-closed；实际连续错误登录触发 429。全局业务限流仍独立保留。
- 生产环境拒绝默认、占位或短于 32 字符的 JWT 密钥；外部 request ID 只接受安全字符；前端导航前会通过 `/auth/me` 复核缓存身份，损坏缓存会被清理。

本轮重新执行 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go build ./...`、`npm run build`、`docker compose config --quiet`、`docker compose up -d --build` 和 `git diff --check`，均退出 0。真实空卷流程完整覆盖读数创建、计划提交/批准、两条执行安排、并发完成、日累计阻断、逆向状态阻断、最终计划闭合及审计行数核对。
