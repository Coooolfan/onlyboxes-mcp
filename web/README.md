# onlyboxes-web

Onlyboxes worker registry dashboard (Vue 3 + Vite + TypeScript).

## 功能

- 首次访问通过账号密码登录控制台（凭据由 `console` 启动时输出）
- 启动后通过 `GET /api/v1/console/session` 完成会话 bootstrap，并基于 `is_admin` 做角色分流
- 管理员默认进入 `/workers`：可查看 worker 控制台、管理 worker、管理自己的 token
- 非管理员默认进入 `/tokens`：仅可管理自己的 token
- `/workers` 路由带管理员守卫，非管理员自动重定向 `/tokens`
- token 管理来自 `GET/POST/DELETE /api/v1/console/tokens`（明文 token 仅在创建响应中返回一次）
- 管理员在 `registration_enabled=true` 时可在 `/workers` 页面创建非管理员账号（`POST /api/v1/console/register`）
- worker 列表来自 `GET /api/v1/workers`
- 统计卡片来自 `GET /api/v1/workers/stats`
- 创建 worker 后自动复制创建响应中的启动命令（明文 `WORKER_SECRET` 仅创建时返回一次）
- `GET /api/v1/workers/:node_id/startup-command` 固定返回 `410 Gone`
- 节点能力列展示 `capabilities[].name` 能力声明
- 支持 `all / online / offline` 筛选、分页、手动刷新和自动刷新

## 开发

```bash
yarn
```

```bash
yarn dev
```

默认开发端口：`5178`

默认开发代理：

- `/api/*` -> `http://127.0.0.1:8089`

可通过环境变量覆盖：

```bash
VITE_API_TARGET=http://127.0.0.1:8089 yarn dev
```

## 构建与测试

```bash
yarn build
yarn test:unit
```
