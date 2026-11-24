# Let's Talk - 实时聊天系统

一个基于 Go 语言和 WebSocket 的实时聊天系统，支持群组聊天功能。

github:[bugoutianzhen123/LetsTalk](https://github.com/bugoutianzhen123/LetsTalk)

## 功能特点

- 实时消息传递
- 群组聊天
- 用户管理
- 并发处理

## 技术栈

- Go 1.21+
- Gin Web 框架
- Gorilla WebSocket
- Zap 日志库

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本（该项目使用 1.23.8）
- MySQL 5.7 或更高版本 （该项目使用 mysql8）

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置数据库

1. 创建 MySQL 数据库
2. 修改 `config/config.yaml` 中的数据库配置

### 4. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 5.测试websocket

文件test/local.html可用于测试websocket

## API 接口

### WebSocket 连接

- 群聊 WebSocket 连接：

```
ws://localhost:8080/ws?groupid=xx&userixx
```

### 消息格式

1. 群聊消息：

```json
{
  "group_id":  1,
  "user_id":   1,
  "content": "Hello everyone!"
}
```

### HTTP 接口

1. 创建用户

```
POST /user/register
//暂于其他功能无关联
```

2. 创建群组

```
POST /group/create
//暂不需要，目前的群聊并不会检测群是否存在数据库，群号相同即可进入同一聊天
```

## 项目结构

```
.
├── config/         # 配置文件
├── domain/         # 领域模型
├── handler/        # HTTP处理器
├── ioc/           # 依赖注入
├── pkg/           # 公共包
├── repository/    # 数据访问层
├── router/        # 路由配置
├── service/       # 业务逻辑层
└── main.go        # 程序入口
```

## 开发说明

1. 代码规范遵循 Go 标准项目布局
2. 使用依赖注入管理组件
3. 采用分层架构设计
4. 使用 WebSocket 实现实时通信

## 注意事项

1. 确保数据库配置正确
2. 检查端口 8080 是否被占用
3. 确保用户 ID 在发送消息时已存在
