请生成 `aquaculture-water-feeding-control`「水产养殖水质与投喂控制」Go 全栈项目，服务规模化水产养殖场。系统根据池塘水质、天气窗口和生长阶段生成投喂建议并管理人工确认，不做电商、库存、记账或预约。

## 项目主要需求

复杂度下限：核心实体不少于 3 个、核心页面不少于 4 个、横切关注点不少于 2 个、共享前端组件不少于 3 个、自定义 hooks/utils 不少于 2 个、后端中间件不少于 2 个。

### 核心实体

`Pond`（养殖池与品种）、`WaterReading`（溶氧/温度/pH 读数）、`FeedingPlan`（投喂策略与版本）、`ControlExecution`（实际执行与反馈）必须从数据库到 Go model/service/handler 再到前端 api/store/page 全链路贯穿。

### 核心页面

`/ponds` 池塘工作台；`/readings` 水质读数与异常确认；`/plans` 投喂计划；`/executions` 执行反馈；`/audit` 操作审计。`RiskTag` 被池塘和读数页共用，`PlanDrawer` 被计划和执行页共用。

### 横切关注点

JWT/RBAC 贯穿角色表、Go 中间件、路由守卫和前端显隐；规则版本审计贯穿计划变更、确认、撤销；全局错误处理中间件、请求 ID 和 Redis 限流必须独立实现。

### 共享枚举/组件

共享 `PondStatus`（active/quarantine/closed）与 `PlanStatus`（draft/pending/approved/executed）。至少提供 `StatusBadge`、`MetricCard`、`ConfirmDialog`，hooks 提供 `useAuth`、`useQueryParams`。

### 技术与规模要求

前端 Vue 3 + TypeScript + Vite + Element Plus；后端 Go 1.22 + Gin + GORM；PostgreSQL + Redis。目标 2800–4000 行 Go 功能代码、28–40 个 `.go` 文件；至少 20 个功能 Go 文件，禁止单文件大杂烩。

### 文件结构强制清单

前端必须有 `api/stores/types/components/common/hooks/pages/router/utils`；后端必须拆分 `model/dto/repository/service/handler/router/middleware/constants/util`。README 列出所有共享枚举位置，并明确每个实体的跨层文件。

### 结构红线

严禁合并职责到单一文件；实体、服务、处理器、路由和前端模块必须保持分层。

### 部署与交付

根目录必须提供 `docker-compose.yml`（顶层 `name: aquaculture-water-feeding-control`，且不写 `version:`）、`.env` 和 `.env.example`（均含 `COMPOSE_PROJECT_NAME=aquaculture-water-feeding-control`）、`README.md`、`frontend/Dockerfile`、`backend/Dockerfile` 和 `frontend/nginx.conf`。前端端口 `18502`、后端端口 `19502`；`/api` 由 Nginx 配置转发到 `backend:8080`，数据库用命名卷和健康检查，后端等待 `condition: service_healthy`。提供真实 `/healthz`，可在中文目录直接 `docker compose up -d`，完成 Git 初始化。
