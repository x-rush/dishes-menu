# 每周菜品安排

> 私人小项目 —— 移动端每周菜品安排 PWA，给女朋友的桌面小组件体验。

> 🚧 **部署文档为主**。本地开发流程见对应章节。

## 特性

- **菜品库** — CRUD 30 道默认家常菜，可自定义新增
- **周菜单** — 周一~周五 (早+加餐) / 周六~周日 (早+午+晚) 自动按规则渲染时段
- **换一换** — 一键随机换菜，内置加权去重（同一天不重样、7 天内降权）
- **筛选** — 口味 / 食材（多选）/ 难度（下拉）
- **双层备注** — 菜品自带做法/小贴士 + 当日点餐备注（少辣、不吃香菜…）
- **PWA** — 可"添加到主屏"离线使用
- **待办** — 共享待办列表,emoji 签名 + 可选截止日期 + 5s 撤销
- **在一起天数** — 心形 FAB 显示已共度天数,点开触发粒子惊喜
- **单容器部署** — 前后端构建产物通过 `//go:embed` 编入 Go 二进制，最终镜像 ~15 MB

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.25 + Gin v1.12 + sqlx + go-sql-driver/mysql |
| 数据库 | MySQL 8（库名 `dishes_menu`） |
| 前端 | Vite 6 + Vue 3.5 + TypeScript + Pinia 3 + vite-plugin-pwa |
| 运行时 | distroless/static:nonroot 镜像 |
| 部署 | Docker Compose（单容器） |

## 项目结构

```
dishes-menu/
├── backend/                      # Go 服务
│   ├── main.go                   # 入口：embed FS + 启动 Gin
│   ├── web/dist/                 # Vite 构建产物（被 //go:embed 编入二进制）
│   ├── migrations/               # 0001_init.{up,down}.sql
│   └── internal/
│       ├── model/                # Dish / WeekMenu 等结构体
│       ├── db/                   # MySQL 连接池
│       ├── dao/                  # SQL CRUD
│       ├── service/              # shuffle 加权算法
│       └── api/                  # Gin 路由 + handlers
├── frontend/                     # Vue 3 PWA
│   ├── src/
│   │   ├── pages/Home.vue        # 单页面
│   │   ├── components/           # WeekTabs / FiltersBar / DishCard / AddDishDialog
│   │   ├── stores/               # Pinia: dishes / menu
│   │   ├── api/client.ts         # fetch 封装
│   │   ├── utils/isoWeek.ts      # ISO 周计算
│   │   └── styles/main.css       # 粉色 / 暖白主题
│   └── public/                   # 静态资源（favicon + PWA 图标）
├── Dockerfile                    # 多阶段构建
├── docker-compose.prod.yml       # 生产部署
├── .env.example                  # 环境变量模板
└── README.md
```

## 快速开始（生产部署）

### 1. 准备 MySQL

```bash
# 任意可访问的 MySQL 8 实例（本地或远程）
mysql -h <host> -u root -p -e \
  "CREATE DATABASE IF NOT EXISTS dishes_menu \
   DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 2. 复制并填写环境变量

```bash
cp .env.example .env
vi .env
# 修改 MYSQL_DSN 为你的真实地址
# 例：MYSQL_DSN=app:secret@tcp(10.0.0.5:3306)/dishes_menu?charset=utf8mb4&parseTime=true&loc=Local
```

### 3. 构建并启动

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

构建约 2-3 分钟（首次需要拉基础镜像）。完成后：

- 浏览器访问 `http://<server-ip>:8080` — 看到 PWA 首页
- 浏览器访问 `http://<server-ip>:8080/api/health` — 看到 `{"status":"ok"}`

### 4. 日常运维

```bash
# 查看日志
docker compose -f docker-compose.prod.yml logs -f

# 停止
docker compose -f docker-compose.prod.yml down

# 升级（拉新代码后）
docker compose -f docker-compose.prod.yml up -d --build
```

## HTTPS 反向代理（可选）

在 80/443 端口前置一层 Caddy（自动 Let's Encrypt）：

```caddyfile
# nginx/Caddyfile
dishes.your-domain.com {
    reverse_proxy localhost:8080
}
```

```bash
# 启动 caddy
docker run -d --name caddy \
  -p 80:80 -p 443:443 \
  -v ./nginx/Caddyfile:/etc/caddy/Caddyfile \
  -v caddy_data:/data \
  -v caddy_config:/config \
  caddy:2
```

## 本地开发

需要：Go 1.22+、Node 22+、可访问的 MySQL 实例。

### 1. 后端

```bash
cd backend
cp ../.env.example .env
# 编辑 .env 填入本地 MYSQL_DSN
go mod download
go run .      # 监听 :8080，首次启动自动跑 migration + 灌种子
```

### 2. 前端

```bash
cd frontend
npm install
npm run dev    # 监听 :5173，/api 自动代理到 :8080
```

### 3. 端到端联调

- 浏览器打开 `http://localhost:5173`
- DevTools → Network → 确认 `/api/dishes`、`/api/menu` 返回 200
- Chrome → Application → Manifest 应显示"每周菜品安排"

## API 速览

```
GET    /api/health
GET    /api/dishes
POST   /api/dishes
PUT    /api/dishes/:id
DELETE /api/dishes/:id
GET    /api/menu?week=2026-W24
PUT    /api/menu/:day/:slot?week=   body: { dish_id, note }
GET    /api/menu/shuffle?week=&day=&slot=&taste=&ingredient=&difficulty=
```

错误信封：`{ "error": { "code": "DISH_NOT_FOUND", "message": "...", "http_status": 404 } }`

## 时段规则

| 星期 | 适用时段 |
|---|---|
| 周一 ~ 周五 | 早餐 + 加餐 |
| 周六 ~ 周日 | 早餐 + 午餐 + 晚餐 |

时段定义见 `backend/internal/model/types.go` 和 `frontend/src/types.ts`。

## 数据库表

```sql
dishes        -- 菜品库（JSON 列存 slots / taste / ingredient）
week_menus    -- 周菜单（扁平结构；uk_week_day_slot 唯一索引）
```

迁移文件 `backend/migrations/0001_init.up.sql`，启动时自动执行（幂等）。

## 验证清单

- [ ] `curl http://localhost:8080/api/health` → `{"status":"ok"}`
- [ ] `curl http://localhost:8080/api/dishes | jq '.dishes | length'` → `30`（首次启动后）
- [ ] 浏览器 → 看到 7 个日 tab、3 个时段卡片（如果是周一~周五）
- [ ] 点"换一换" → 卡片切换 + 数据库 `week_menus` 表新增/更新一行
- [ ] 点"备注" → 弹层 + 保存后刷新仍在
- [ ] Chrome DevTools → Lighthouse → PWA 类别 ≥ 90 分
- [ ] 手机浏览器 → "添加到主屏" 出现粉色图标

## 已知限制

- 无鉴权（私密 URL 部署；不暴露公网）
- 无并发写入冲突检测（单用户场景不需要）
- Service worker 只缓存 GET 请求（API PUT/POST 不走 SW）

## License

私人项目。
