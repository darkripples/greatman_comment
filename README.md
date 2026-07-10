
# greatman_comment 用 AI 重新看见人 · 人文季 Demo

接入知乎热榜与历史人物agent，以历史人物的语气点评当前热榜新闻

历史人物 × 知乎热榜对话。前后端分离：

- `server/` — Go 后端（热榜/搜索走知乎，对话支持知乎直答与 DeepSeek 切换；支持单人对话与多人群聊）
- `web/` — Next.js 15 响应式前端（页面可配置 API 环境与 LLM 提供方）

## 快速启动

根目录双击 **`start.bat`**，或分别启动后端与前端。

### 1. 环境变量（Go 后端）

在系统环境变量或终端中设置（**仅需 API Key**）：

```powershell
$env:ZHIHU_API_KEY = "..."          # 知乎热榜/搜索（及直答 LLM）
$env:DEEPSEEK_API_KEY = "sk-..."      # DeepSeek 对话
```

其余配置（LLM 提供方、缓存间隔、Mock、API 地址等）在页面 **设置** 中修改，保存至 SQLite。

可选启动参数（非敏感，一般不必改）：

```powershell
# $env:SERVER_PORT = "30302"
# $env:RENWEN_DATA_DIR = "...\server\data"
```

### 2. 启动 Go 后端

```powershell
cd server
go run ./cmd/server
# 默认 http://127.0.0.1:30302
```

### 3. 启动 Next 前端

```powershell
cd web
copy .env.local.example .env.local
npm install
npm run dev
# http://localhost:30301
```

## 前端页面设置

打开页面右上角 **「设置」** 抽屉：

| 选项 | 说明 |
|------|------|
| API 环境 · 本地 Dev | 开发阶段使用 |
| 本地模式 · Next 反代 | 请求 `/api/*`，由 `next.config.ts` 转发到 `:30302`（推荐） |
| 本地模式 · 直连 Go | 浏览器直接请求 `http://127.0.0.1:30302/api/*` |
| API 环境 · 线上 Prod | 填写线上 API 根地址（如 `https://api.example.com`） |
| LLM 提供方 | 选择 `DeepSeek` 或 `知乎直答` |
| 高级 | Mock 热榜、缓存 TTL、最小拉取间隔、模型名等 |

设置通过 `GET/PUT /api/settings` 读写，持久化在 SQLite `app_settings` 表；**环境变量仅保留 API Key**。

## 对话模式

| 模式 | 说明 |
|------|------|
| 单人 | 选一位历史人物，基于热榜议题提问 |
| 群聊 | 选 2–3 位人物，逐人串行 LLM 发言，支持多轮辩论 |

群聊时后端会按人物顺序依次调用 LLM，每位人物能看到此前发言摘要，形成「圆桌讨论」效果。

## 数据流与持久化（SQLite）

**原则：页面只请求本 Go 后端；后端对知乎热榜/搜索先查库、按需拉取、入库后再返回库中数据。LLM 对话除外，仍实时调用模型 API。**

数据库文件：**`{RENWEN_DATA_DIR}/renwen.db`**（默认 `server/data/renwen.db`）

| 数据 | 存储 | 说明 |
|------|------|------|
| 热榜 / 搜索 | `api_cache` 表 | 先读 SQLite；无缓存或过期且超过最小间隔才请求知乎，写入后再从库返回 |
| 对话 / 群聊 | `conversations` + `messages` | 持久化历史，供 `/api/conversations` 查询 |

### 热榜/搜索缓存间隔

在页面 **设置 → 高级** 中配置（写入 SQLite），例如：

- 热榜/搜索缓存 TTL（默认 **5 分钟**）
- 热榜最小拉取间隔（默认 **5 小时**，避免频繁消耗知乎日配额）
- 搜索最小拉取间隔（默认 **5 小时**）

开发阶段建议保持较长间隔（`5h`~`24h`），或开启「知乎 Mock」使用本地 fixtures。也可通过 `GET/PUT /api/settings` 读写。

热榜 API 响应字段：`cached`（命中库）、`stale`（间隔内返回旧数据）、`source: "sqlite"`、`nextFetchAt`（下次可拉取时间戳）。

首次启动若存在旧版 `renwen.json`，会自动导入到 SQLite。

## API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| GET | `/api/settings` | 读取应用设置（SQLite） |
| PUT | `/api/settings` | 更新应用设置（SQLite） |
| GET | `/api/hot-list` | 热榜（**只读 SQLite**；按需拉取知乎入库后回显，`cached`/`stale`/`source`） |
| GET | `/api/search?q=` | 搜索（同上） |
| GET | `/api/characters` | 历史人物列表 |
| GET | `/api/providers` | 可用 LLM 提供方 |
| POST | `/api/chat` | 单人对话 `{ characterId, question, conversationId?, sourceTitle?, provider? }` |
| POST | `/api/group-discuss` | 群聊 `{ characterIds, question, history, round, conversationId?, ... }` |
| GET | `/api/conversations` | 对话列表 |
| GET | `/api/conversations/{id}` | 对话详情与消息 |

## 目录结构

```
renwen/
├── start.bat / stop.bat
├── server/
│   ├── cmd/server/main.go
│   ├── config/characters.json
│   ├── config/group_context.txt
│   ├── data/renwen.db            # SQLite（运行时生成）
│   ├── data/renwen.json          # 旧版数据，首次可自动导入
│   ├── fixtures/hot_list.json
│   └── internal/
│       ├── storage/store.go      # SQLite 持久化
│       ├── discussion/orchestrator.go
│       └── ...
└── web/
    ├── app/page.tsx
    ├── components/
    └── lib/{api,settings,types}.ts
```

## 人文季投稿提醒

1. AI Works 上传项目
2. 文章标题前缀：`【人文季-历史单元投稿】`
3. 话题：`#用AI重新看见人` `#用AI重访一段历史`
