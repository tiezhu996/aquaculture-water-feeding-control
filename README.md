# 水产养殖水质与投喂控制

`aquaculture-water-feeding-control` 是面向规模化水产养殖场的 Go 全栈作业系统。它把养殖池、水质读数、投喂计划和现场执行连成可审计流程，并根据最新水质、天气窗口和生长阶段生成投喂建议。

## 功能范围

- 养殖池工作台：管理品种、容量、生长阶段和运行/隔离/关闭状态。
- 水质风险：录入溶解氧、水温、pH、氨氮和浊度，自动判定正常/预警/严重并要求人工确认异常。
- 投喂计划：草稿修订自动升版，支持提交、批准和撤销；批准时会校验水质与溶解氧阈值。
- 投喂建议：结合已批准计划、24 小时内水质、天气和生长阶段，输出正常投喂、减量或暂停。
- 执行反馈：仅允许已批准计划进入执行，记录实际量、现场溶解氧与反馈。
- 安全与审计：JWT、RBAC、请求 ID、全局异常恢复、Redis 限流和实体变更前后快照。

## 快速启动

需要 Docker 24+ 和 Docker Compose v2+。项目路径可包含中文。

```bash
cp .env.example .env
# 生产环境请修改 POSTGRES_PASSWORD 和 JWT_SECRET
docker compose up -d --build
docker compose ps
```

- 前端：<http://localhost:18502>
- 后端：<http://localhost:19502>
- 健康检查：<http://localhost:19502/healthz>
- Nginx 会将前端同源 `/api` 请求转发到 `backend:8080`。

停止服务（保留 PostgreSQL 和 Redis 命名卷）：

```bash
docker compose down
```

如需同时删除本项目的开发数据，明确执行 `docker compose down -v`。

## 演示账号

| 角色 | 用户名 | 密码 | 主要权限 |
| --- | --- | --- | --- |
| 管理员 | `admin` | `admin123` | 全部功能 |
| 生产主管 | `manager` | `manager123` | 养殖池管理、计划审核、审计查看 |
| 值班操作员 | `operator` | `operator123` | 水质录入、计划草稿、执行反馈 |
| 观察员 | `viewer` | `viewer123` | 只读业务数据 |

这些账号由后端首次启动时幂等创建，密码以 bcrypt 存储。正式部署时应替换为组织身份源或修改初始密码。

## 核心操作流程

1. 在“养殖池”创建或选择一个 `active` 养殖池。
2. 在“水质读数”录入当前指标；如判定异常，先进行现场复核和确认。
3. 在“投喂计划”创建草稿并提交，主管检查最新水质后批准。
4. 已批准计划可输入天气窗口生成实时投喂建议。
5. 在“执行反馈”安排、开始并提交实际结果。完成后计划进入 `executed`。
6. 管理员或主管可在“操作审计”查看人员、原因、请求 ID 和变更快照。

## 项目结构

```text
.
├── backend/
│   ├── cmd/api/                 # 进程入口与依赖装配
│   └── internal/
│       ├── config/ database/     # 配置、迁移与种子数据
│       ├── constants/ model/     # 共享枚举与 GORM 实体
│       ├── dto/ repository/      # API 输入输出与数据访问
│       ├── service/ handler/     # 业务规则与 HTTP 处理
│       └── middleware/ router/   # JWT/RBAC/限流/请求 ID 与路由
├── frontend/src/
│   ├── api/ stores/ types/       # 请求层、Pinia 状态与类型
│   ├── components/common/       # 共享工作台组件
│   ├── hooks/ router/ utils/     # 权限、查询参数、守卫和工具
│   └── pages/                   # 五个业务页与登录页
└── docker-compose.yml
```

### 共享枚举位置

- Go：`backend/internal/constants/enums.go`
  - `PondStatus`: `active` / `quarantine` / `closed`
  - `PlanStatus`: `draft` / `pending` / `approved` / `executed`
  - `RiskLevel`、`ExecutionStatus`、`Role`
- TypeScript：`frontend/src/types/enums.ts`
  - 与 Go 取值一致，同时提供界面文案映射。

### 核心实体跨层文件

| 实体 | Model | DTO | Repository | Service | Handler | Frontend API | Page |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `Pond` | `model/pond.go` | `dto/pond.go` | `repository/pond_repository.go` | `service/pond_service.go` | `handler/pond_handler.go` | `api/ponds.ts` | `pages/PondsPage.vue` |
| `WaterReading` | `model/water_reading.go` | `dto/reading.go` | `repository/reading_repository.go` | `service/reading_service.go` | `handler/reading_handler.go` | `api/readings.ts` | `pages/ReadingsPage.vue` |
| `FeedingPlan` | `model/feeding_plan.go` | `dto/plan.go` | `repository/plan_repository.go` | `service/plan_service.go` | `handler/plan_handler.go` | `api/plans.ts` | `pages/PlansPage.vue` |
| `ControlExecution` | `model/control_execution.go` | `dto/execution.go` | `repository/execution_repository.go` | `service/execution_service.go` | `handler/execution_handler.go` | `api/executions.ts` | `pages/ExecutionsPage.vue` |

`RiskTag` 在养殖池和水质页共用，`PlanDrawer` 在计划和执行页共用。`StatusBadge`、`MetricCard`、`ConfirmDialog` 位于 `frontend/src/components/common/`；`useAuth`、`useQueryParams` 位于 `frontend/src/hooks/`。

## API 摘要

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/auth/login` | 登录并获取 JWT |
| `GET/POST` | `/api/ponds` | 养殖池列表/创建 |
| `GET/POST` | `/api/readings` | 水质列表/录入与风险判定 |
| `PATCH` | `/api/readings/:id/confirm` | 确认异常读数 |
| `GET/POST` | `/api/plans` | 投喂计划列表/创建 |
| `PATCH` | `/api/plans/:id/submit` | 草稿提交审核 |
| `PATCH` | `/api/plans/:id/approve` | 主管批准并校验水质 |
| `PATCH` | `/api/plans/:id/revoke` | 撤销待审或已批准计划 |
| `GET` | `/api/plans/recommendation?pondId=1&weather=晴朗` | 生成投喂建议 |
| `GET/POST` | `/api/executions` | 执行记录列表/安排 |
| `PATCH` | `/api/executions/:id/complete` | 提交实际数量与反馈 |
| `GET` | `/api/audit` | 管理员/主管查看审计记录 |

错误统一为 `{"error":{"code":"...","message":"...","requestId":"..."}}`，响应头同时包含 `X-Request-ID`。

## 本地开发与验证

后端（需要可用的 PostgreSQL 和 Redis）：

```bash
cd backend
gofmt -w .
go test ./...
go build ./cmd/api
```

前端：

```bash
cd frontend
npm install
npm run build
npm run dev
```

Compose 静态检查：

```bash
docker compose config --quiet
```

## 业务约束

- 关闭养殖池不接受新水质读数或投喂计划。
- 草稿每次编辑版本号加一；非草稿不能直接编辑。
- 计划批准需要运行中养殖池和最新水质，溶解氧不得低于计划阈值。
- 执行安排需要 24 小时内水质，严重异常或溶解氧不足会阻断流程。
- 实际量与计划量偏差超过 25% 时，必须提供至少 10 个字的说明。
- 关联了读数、计划或执行记录的养殖池不允许删除。
