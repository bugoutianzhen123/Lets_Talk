# LetsTalk · 网页版实时聊天系统

一个前后端分离的即时通讯（IM）应用：注册登录、好友申请、一对一私聊、建群/加群、群聊广播，全部基于 WebSocket 实时收发，消息落库可查历史。

- 后端：Go + Gin + GORM，WebSocket 长连接，MySQL 持久化
- 前端：React + TypeScript + Vite + Ant Design，Zustand 管理登录态

---

## 一、功能特性

| 模块 | 功能 |
|------|------|
| 账号 | 注册、登录（用户名唯一） |
| 好友 | 搜索用户、发起申请、同意申请、好友列表、在线状态 |
| 私聊 | 一对一实时聊天、聊天历史、未读消息角标 |
| 群聊 | 创建群聊、搜索并加入群聊、群内多人实时广播、群历史 |
| 会话 | 会话按最近消息排序、消息预览、未读红点、自动滚动到底部 |
| 其他 | 本地生成头像（无需联网）、全局中文界面、接口跨域支持 |

---

## 二、技术栈

### 前端

| 类别 | 选型 |
|------|------|
| 框架 | React 18（函数组件 + Hooks） |
| 语言 | TypeScript 5 |
| 构建 | Vite 5 |
| UI 组件 | Ant Design 6 + @ant-design/icons |
| 路由 | react-router-dom 7 |
| 状态管理 | Zustand 5（`persist` 中间件持久化到 sessionStorage） |
| 网络 / 实时 | 原生 `fetch` + 原生 `WebSocket` |
| 时间处理 | moment |

### 后端

| 类别 | 选型 |
|------|------|
| 语言 | Go |
| Web 框架 | Gin |
| ORM | GORM（MySQL 驱动，自动建库建表 + 种子数据） |
| WebSocket | gorilla/websocket |
| 配置 | Viper（读取 `config/config.yaml`），pflag 指定配置文件 |
| 日志 | zap + lumberjack |
| 其他 | gin-contrib/cors（跨域）、gin-contrib/pprof（性能分析） |

---

## 三、项目结构

```
Lets_Talk/
├── be/                     # 后端（Go）
│   ├── config/             # 配置文件 config.yaml
│   ├── domain/             # 领域模型（用户、好友、群、聊天记录等）
│   ├── handler/            # HTTP / WebSocket 处理器
│   ├── ioc/                # 依赖注入（DB、日志、Redis）
│   ├── pkg/logger/         # 日志封装
│   ├── repository/         # 数据访问层（repository / dao / cache）
│   ├── router/             # 路由注册
│   ├── service/            # 业务逻辑 + WebSocket 连接/广播管理
│   ├── test/local.html     # WebSocket 测试页
│   ├── Dockerfile
│   └── main.go             # 程序入口
├── fe/                     # 前端（React + Vite）
│   ├── src/
│   │   ├── pages/          # Login 登录页、ChatLayout 聊天主界面
│   │   ├── store/          # Zustand 全局状态
│   │   ├── App.tsx         # 路由与全局配置
│   │   └── main.tsx        # 入口
│   ├── vite.config.ts      # 开发代理 + 预览代理
│   └── Dockerfile
├── docker-compose.yml      # 一键容器化部署
├── start_be.sh / .bat      # 后端启动脚本
└── start_fe.sh / .bat      # 前端启动脚本
```

---

## 四、快速开始

### 1. 环境要求

- Go（建议使用 `be/go.mod` 中声明的版本；本地验证使用 Go 1.22）
- Node.js 18+ 与 npm
- MySQL 8（本地默认 `127.0.0.1:3306`）

### 2. 准备数据库

后端启动时会**自动创建数据库和表**，并写入种子数据（用户权限、申请状态、群身份，以及 3 个演示群聊），无需手工执行建表 SQL。

只需确保 MySQL 已启动，并在 `be/config/config.yaml` 中填好连接信息：

```yaml
mysql:
  dsn: "root:123456@tcp(127.0.0.1:3306)/lets_talk?charset=utf8mb4&parseTime=True&loc=Local"
```

### 3. 启动后端

```bash
cd be
go mod tidy
go run main.go
# 服务监听 http://localhost:8080
```

或使用脚本：`./start_be.sh`（Windows 用 `start_be.bat`）。

### 4. 启动前端

```bash
cd fe
npm install
npm run dev
# 开发服务器 http://localhost:5173
```

或使用脚本：`./start_fe.sh`（Windows 用 `start_fe.bat`）。

前端所有请求使用**相对路径**，开发环境由 Vite `proxy` 转发到 `http://localhost:8080`，因此无需在代码里配置后端地址。

### 5. 生产构建与预览

演示时建议用生产构建，首屏更快、更稳定（不再现场编译）：

```bash
cd fe
npm run build     # 产出 fe/dist 静态文件
npm run preview   # 以静态方式提供，默认 http://localhost:5173
```

> 注意：`preview` 提供的是打包产物，修改前端代码后需重新 `npm run build`。

### 6. Docker 一键部署

```bash
docker compose up -d --build
```

- 前端（Nginx）：宿主 `3000` → 容器 `80`
- 后端：宿主 `8080` → 容器 `8080`

> `fe/Dockerfile` 中 Nginx 默认未配置反向代理，如需前后端联调，请在 Nginx 中把 `/user`、`/friend`、`/group`、`/ws` 反代到后端 `8080`。

---

## 五、接口文档

### HTTP 接口

所有响应统一为 JSON，成功时业务数据放在 `data` 字段。

| 方法 | 路径 | 说明 | 主要参数 |
|------|------|------|----------|
| POST | `/user/register` | 注册 | body: `{ username, password }` |
| POST | `/user/login` | 登录 | body: `{ username, password }` → 返回 `{ user: { id, username } }` |
| GET | `/user/search` | 搜索用户 | query: `keyword` |
| POST | `/friend/apply` | 发起好友申请 | body: `{ sender_id, receiver_id, message }` |
| GET | `/friend/applications` | 收到的好友申请 | query: `userid` |
| POST | `/friend/accept` | 同意好友申请 | body: `{ application_id }` |
| GET | `/friend/list` | 好友列表（含在线状态、最后一条消息） | query: `userid` |
| GET | `/friend/history` | 与某好友的聊天记录 | query: `userid`, `friendid` |
| GET | `/group/list` | 群聊列表 | query: `userid`（可选，用于标记 joined） |
| POST | `/group/create` | 创建群聊 | body: `{ name, creator_id }` |
| POST | `/group/join` | 加入群聊 | body: `{ group_id, user_id }` |
| GET | `/group/history` | 群聊历史记录 | query: `groupid` |

### WebSocket 接口

| 路径 | 说明 | 参数 |
|------|------|------|
| `/ws?groupid={id}&userid={id}` | 群聊实时消息（连接即进入该群） | 查询参数 |
| `/ws/friend?userid={id}` | 私聊通道（每个用户常驻一条，负责收发私聊） | 查询参数 |

- **群聊消息**：直接发送纯文本，服务端补齐发送人信息后广播。
- **私聊消息**：客户端发送 JSON `{ "target_id": 2, "content": "hi" }`；服务端推送 `{ sender_id, username, target_id, content, send_time }`。

> 消息一律**先写入数据库再推送**，因此对方离线时重新打开会话可从历史接口拉回。

---

## 六、数据库表

由 GORM `AutoMigrate` 自动创建：`users`、`user_permissions`、`friends`、`friend_chat_histories`、`friend_applications`、`request_statuses`、`groups`、`group_identities`、`group_members`、`group_chat_histories`。

初始化时会写入种子数据：用户权限（user / super_user）、申请状态（申请中 / 已同意 / 已拒绝）、群身份（群主 / 管理员 / 一般群员），以及 3 个演示群聊。

---

## 七、常见问题

**1. 前端首屏加载慢或偶发报错？**
`npm run dev` 是开发服务器，首次访问需现场编译大量模块。演示请改用 `npm run build && npm run preview`。

**2. 服务器上运行 `node` / `mysql` 报 `GLIBCXX_3.4.xx not found`？**
运行环境注入了旧版 `libstdc++`（如 IDE 的 `LD_LIBRARY_PATH`）。启动前执行 `unset LD_LIBRARY_PATH` 即可。

**3. 用公网 IP 打不开前端？**
Vite 默认只监听 `127.0.0.1`。`vite.config.ts` 已配置 `server.host = true` / `preview.host = true` 监听 `0.0.0.0`，同时确认云服务器安全组已放行对应端口（前端 5173、后端 8080）。

**4. 同一浏览器如何演示两个账号？**
登录态存放在 `sessionStorage`，**新开两个标签页**分别登录不同账号即可（不要使用"复制标签页"）。

**5. 端口被占用？**
后端默认 `8080`、前端开发 `5173`，请确认端口空闲，或调整 `main.go` 中的监听端口 / Vite 配置。

---

## 八、说明

- 前端接口全部使用相对路径，一套代码适配开发代理与线上 Nginx 反代。
- 头像按用户名哈希本地生成 SVG，不依赖外部图片服务。
- 本项目为教学 / 演示用途，账号密码为明文存储，未接入鉴权 Token，请勿用于生产环境。
