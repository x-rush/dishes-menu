# 待办事项 + 底部 tab 栏 + 纪念日惊喜

- **Date**: 2026-06-15
- **Status**: Design (awaiting review)
- **Target branch**: `master`
- **Replaces**: 现有单页 `Home.vue`(Home.vue 保留,改为 `/menu/:week/:date` 路由)

## 1. 目标 & 非目标

### 1.1 目标

为已有「今天想吃什么呀」PWA 增加**待办事项**能力,并通过**底部 tab 栏**让「菜单 / 待办」在同一个 app 内切换;**tab 栏中央的「心 + 在一起天数」** 是给女友的核心情感入口。

### 1.2 非目标(YAGNI)

- ❌ 用户鉴权 / 登录态(单用户私人 PWA)
- ❌ PWA Web Push 通知(单设备场景工程量爆炸,后续若有需要再追加)
- ❌ 重复任务(每天/每周,共享情感场景用得少)
- ❌ 自定义列表(Things 风),只用平铺 + 时间小角标
- ❌ 共享「小纸条墙」暖话抽屉(留作未来扩展,本设计只做静态心 + 天数)
- ❌ 待办的子任务 / 评论 / 附件
- ❌ 待办的导出/导入
- ❌ 实时同步(WebSocket / SSE);**首次切 tab 时 re-fetch** 即可

## 2. 核心产品决策(brainstorming 锁定)

| # | 维度 | 决策 |
|---|------|------|
| 1 | 共享语义 | 双向添加、双方可勾 |
| 2 | 完成行为 | 勾掉 → 折成灰色划掉(5 秒撤销) |
| 3 | 组织方式 | 无分类平铺,只标"过期/今天/未来"小角标 |
| 4 | 截止日期 | 可选(无 / 今天 / 明天 / 周末 / 选一天) |
| 5 | 创建者标识 | emoji 签名徽章(无需身份选择) |
| 6 | tab 栏 | 双 tab 浮起 + 中央心形 |
| 7 | 中央心形 | 静态心 + 在一起天数 + 点击微动 |
| 8 | 后端策略 | 复用 Go 服务 + 新增 `todos` + `counters` 表 |

## 3. 架构总览

```
                    ┌────────────────────────────────┐
                    │      前端 (Vue 3 SPA)         │
                    │  ┌──────────────────────────┐  │
                    │  │ App.vue                  │  │
                    │  │  ├ <RouterView>          │  │
                    │  │  │   ├ /menu  → Home.vue  │  │  ← 现有(几乎不动)
                    │  │  │   └ /todo  → TodoPage  │  │  ← 新增
                    │  │  └ <TabBar>  (菜单|♡|待办)│  │  ← 新增
                    │  │  └ <ThemeToggle/UndoToast>│ │
                    │  └──────────────────────────┘  │
                    │   composables/useTogether.ts (新)│
                    │   stores/todo.ts (新增)         │
                    └──────────────┬─────────────────┘
                                   │  /api/todos, /api/together
                                   ▼
                    ┌────────────────────────────────┐
                    │   Go 后端 (复用 + 增量)         │
                    │   internal/dao/todo.go  (新)   │
                    │   internal/api/handlers_todo.go│
                    │   migrations/0005_*.up.sql (新) │
                    └──────────────┬─────────────────┘
                                   │
                                   ▼
                    ┌────────────────────────────────┐
                    │  MySQL  dishes_menu 库          │
                    │  现有:dishes / week_menus       │
                    │  新增:todos / counters          │
                    └────────────────────────────────┘
```

**关键约束**:
- **TabBar** 是 App 级的浮层,不属于任何路由 —— 菜单页内部的 `/:week/:date` URL 仍可独立工作
- 两个 store 互不感知;tab 切换不卸载对方组件
- **不做乐观更新** —— 完成/添加 全部等后端 200 才生效(避免"勾了但 DB 失败"的她以为完成)
- 后端 embed 的 `web/dist` 增量 < 500KB(纯 Go + 静态资源)

## 4. 数据模型 — migration 0005

```sql
-- 0005_todo_and_counter.up.sql
-- 待办清单(扁平表,与 week_menus 同构)
CREATE TABLE todos (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  content       VARCHAR(500) NOT NULL,
  due_date      DATE         NULL,               -- 可选
  author_emoji  VARCHAR(8)   NOT NULL,           -- e.g. "🌸"
  author_color  VARCHAR(16)  NOT NULL,           -- e.g. "#f8a5c2"
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at  DATETIME     NULL,               -- null = 未完成
  -- 不需要 updated_at:本表更新只有「完成/取消完成」,可从 completed_at 推断
  INDEX idx_todos_open (completed_at, due_date, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通用键值表,存纪念日
CREATE TABLE counters (
  name  VARCHAR(64)  NOT NULL PRIMARY KEY,
  value VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- 灌入一行:name='together_since', value='2024-03-15' (具体日期由 ops 在 release 前手动 SET)
```

**初始数据**:counter 行**不在 migration 里 INSERT**(避免敏感日期被 git 留下),改为部署后 ops 步骤手动 `INSERT INTO counters VALUES ('together_since', '2024-03-15');`(由 deploy 文档说明)。

**索引选择说明**:
- 主查询 `WHERE completed_at IS NULL ORDER BY due_date ASC, created_at ASC` 命中 `idx_todos_open`(leading column = completed_at,range scan of `IS NULL`,then sort by due_date)
- 不需要"按作者查询" → 暂不加 author 索引

## 5. 路由 & 状态

### 5.1 路由表(新增/调整)

| path | component | 说明 |
|------|-----------|------|
| `/` | — | redirect 到 `/menu/2026-W24/今天`(本周今天) |
| `/menu/:week/:date` | `Home.vue` | 现有,完全不动 |
| `/menu/:week` | — | redirect 到 `/menu/:week/今天` |
| `/todo` | `TodoPage.vue` | **新** |
| `*` | — | redirect 到 `/menu` |

**关键点**:`/menu/*` 是新加的命名空间,旧 URL `/:week/:date` 自动失效 —— **主动失效**。理由:现有 URL 没在任何外部系统引用(私人 PWA,无 SEO 需求),命名空间化更清晰,旧用户在 PWA 内会被一次跳到 `/menu/2026-W24/今天`,不影响。

### 5.2 store 拆分

| store | 职责 | 复用 |
|-------|------|------|
| `useMenuStore`(现有) | dishes / week_menus | 不动 |
| `useTodoStore`(新) | todos list + 操作 | setup 风格,跟 useMenuStore 对称 |
| `useTogether`(composable) | 读纪念日 + 算天数,自带 1h 内存缓存 | 不入 store,因为是只读 + 跨页面共享 |
| `useEmoji` (composable,可省) | 从 localStorage 读"上次选用的 emoji",省得每次都点 | 仅在 TodoInputBar 用 |

### 5.3 路由同步

- `App.vue` 持有 `<TabBar/>`,TabBar 内部用 `useRoute()` 高亮当前 tab,点击触发 `router.push`
- **不监听路由变化 fetch 数据** —— 改由各页面 mount 时自己 load(沿用 Home.vue 的模式,避免 tab 切换时所有 store 重 fetch)

## 6. 组件 & composables 清单

### 6.1 新增

```
frontend/src/
  pages/TodoPage.vue                   # 待办主页:hero + 列表 + 输入栏
  components/
    TabBar.vue                        # 双 tab 浮起 + 中央心形胶囊
    TodayTogether.vue                  # 天数庆祝浮层(点击中央心形触发)
    TodoInputBar.vue                  # 顶部快速添加:输入框 + emoji + 日期
    TodoCard.vue                      # 单条待办卡片
    DateChipBar.vue                   # 快捷日期(今天/明天/周末/选)
    EmojiPicker.vue                   # 12 个 emoji 网格 + 8 个预设色
  composables/
    useTogether.ts                    # 读 /api/together + 算 days + 算 next milestone
  stores/todo.ts                      # todos CRUD + 撤销 undo push
  api/client.ts                       # 增 6 个方法:
                                       #   listTodos / createTodo / patchTodo / deleteTodo
                                       #   getTogether / setTogether
```

### 6.2 复用(完整清单)

| 既有 | 复用点 |
|------|--------|
| `useUndo.ts` | 完成/删除/加待办的撤销(单槽语义,5s 过期) |
| `useTheme.ts` | 中央心形的"心"在暗色模式颜色用 `var(--color-pink-400)` 自动适配 |
| `SkeletonCard.vue` | TodoPage 初次加载占位 |
| `ConfettiBurst.vue` | 整百/整千天撒花 + 浮层出现 |
| `<RouterView/>` | 路由出口 |
| `@vueuse/core` 的 `useStorage` | emoji 签名持久化(她下次进来默认用上次选的) |
| `api/client.ts` 的 BASE 自适配 | dev/prod 双模式继续生效 |
| `styles/main.css` 的 design token | 心形/卡片色直接用 `var(--color-pink-*)` |

## 7. 后端 API 详细

### 7.1 路由(在现有 `api.RegisterRoutes` 中追加)

```
GET    /api/todos
POST   /api/todos
PATCH  /api/todos/:id
DELETE /api/todos/:id
GET    /api/together
POST   /api/together
```

### 7.2 行为规约

#### `GET /api/todos`
- 返回 `{todos: Todo[]}`
- 后端排序:`completed_at IS NOT NULL` 沉底,然后 `due_date IS NULL` 沉底,然后 `due_date ASC, created_at ASC`
- 包含全部(已完成 + 未完成),由前端决定"隐藏已完成"开关

#### `POST /api/todos`
- body: `{content, due_date?, author_emoji, author_color}`
- 校验:content 1..500 字符;author_emoji 非空且长度 ≤ 8;author_color 形如 `#xxxxxx`(允许 hex / 预定义色名)
- 响应:201 + 完整 Todo

#### `PATCH /api/todos/:id`
- body: 至少一个 `{content?, due_date?, completed?}`
- `completed: true` 时设 `completed_at = NOW()`;`completed: false` 时设 `completed_at = NULL`
- 响应:200 + 完整 Todo

#### `DELETE /api/todos/:id`
- 204 No Content

#### `GET /api/together`
- 响应:`{since: "2024-03-15", days: 487, next_milestone: 500, days_to_next: 13}`
- 若 counter 不存在:`{since: null, days: 0, next_milestone: 100, days_to_next: null}`(前端用此态显示"心灰,点我设纪念日")

#### `POST /api/together`
- body: `{since: "2024-03-15"}`(YYYY-MM-DD 严格校验)
- 行为:**幂等** — 已存在则忽略,首次则 INSERT
- 响应:200 + 跟 GET 同样的 shape
- **限流**: 同一分钟内多次调用 → 第一次写入,后续返回当前值(避免"误操作改纪念日")

### 7.3 错误码(沿用 `api/error.go` 信封)

| HTTP | code | 触发条件 |
|------|------|----------|
| 400 | `BAD_REQUEST` | content 空、author_emoji 缺失、日期格式错 |
| 404 | `NOT_FOUND` | PATCH/DELETE 不存在的 id |
| 422 | `INVALID_DATE` | since 不是 YYYY-MM-DD 或未来日期 |
| 500 | `INTERNAL` | DB 错误 |

## 8. 视觉 & 交互细节

### 8.1 TabBar

```
┌──────────────────────────────────┐
│  🍚 菜单    💗12    📋 待办      │  ← 底部 safe-area
│          (HeartDays: 487)        │
└──────────────────────────────────┘
         浮起中央胶囊,粉色双阴影
```

- 背景:`var(--color-warm-bg)` + `backdrop-filter: blur(20px)` 玻璃感
- 高度:56px(不含 safe area)
- 中央心形胶囊:直径 64px,顶部凸出 8px
- 选中态:图标 `var(--color-pink-500)` + 文字加粗
- 暗色模式:背景 `var(--color-cream)`,心形颜色自动切到 `var(--color-pink-400)`

### 8.2 中央心形(`useTogether` + `TodayTogether`)

- **常驻状态**:常亮粉色,右上角气泡显示 `days`(整数)
- **点击**:弹出 `TodayTogether` 浮层(底部 sheet 风,跟 DishPickerSheet 同款)
  - 大字"我们在一起的第 487 天"
  - 3 张插画(草图,这次只做 SVG 占位)
  - "距离 500 天还有 13 天"小字
  - "关闭"按钮
- **整百/整千天(100/365/500/1000/10000)**:
  - 首次进入 PWA 时(若当天是 milestone 且未庆祝过)→ 撒 ConfettiBurst + 弹全屏庆祝卡
  - 用 `localStorage.dishes-menu-last-celebrated-milestone = "500"` 记录
- **纪念日未设**:心形灰(`var(--color-muted)`),点 → 弹"点这里设纪念日"小卡 → 输入日期

### 8.3 待办卡片

```
┌──────────────────────────────────┐
│🌸 [ ] 买菜                        │  ← 4px 粉色色条(作者色)+ 圆形 emoji
│         今天 18:30                │
└──────────────────────────────────┘
```

- 左侧 4px 色条 = `author_color`
- checkbox:点击切换完成
- 内容:可点击编辑(长按弹小输入框)
- 时间角标颜色:
  - 过期 = 红色"已过期 2 天"
  - 今天 = 蜜桃色"今天"
  - 明天 = 粉色"明天"
  - 本周 = 粉色"周X"
  - 更远 = 灰色"M月D日"
  - 无 = 灰色"不急"
- 勾掉:整条 `opacity: 0.5`,`text-decoration: line-through`,`font-size` 降 1px,保留 5s 可撤销
- 删除:右上角小 × 按钮(点击不立即删,先弹撤销 toast,5s 后才真删)

### 8.4 列表分组

**有日期区**(顶部):
```
[今天]   (1)
  🌸 [ ] 买菜
[明天]   (2)
  🐱 [ ] 还信用卡
  🌸 [ ] 给她发消息
[本周]   (1)
  ...
[更远]   (1)
  ...
[不急]   (3)  ← 可折叠
```

### 8.5 空状态

无任何待办时:中间一只蒸着热气的盘子插画(复用 `EmptyPlate`?或新画一只) + "还没添加待办,试试在下面加一条" + 居中提示箭头

## 9. 错误处理

| 场景 | 行为 |
|------|------|
| 创建/完成 500 | toast 红色,操作回滚(已完成的乐观更新 — 需 200 才勾) |
| 网络断开(SW 缓存的 GET 可用) | 列表正常,POST/PATCH 报"网络异常" |
| 纪念日未设 | 心形灰,提示"点这里设纪念日" |
| 撤销过期 | toast 自动消失,不可点"撤销" |

## 10. 验收清单

- [ ] 加一条无日期待办 → 出现在"不急"折叠区
- [ ] 加一条今天待办 → 顶部"今天"区
- [ ] 加一条明天待办 → "明天"区
- [ ] 勾掉 → 折灰 + 5s 撤销 toast
- [ ] 5s 内点"撤销" → 恢复未完成
- [ ] 切到菜单 tab → 菜单数据完整;切回 → 待办数据完整(无 re-fetch 闪烁)
- [ ] 中央心形显示"第 N 天"(假设纪念日 2024-03-15)
- [ ] 点击心形 → 弹 TodayTogether 浮层
- [ ] 纪念日未设 → 心形灰 + 提示
- [ ] 删除最后一条 → 空状态插画
- [ ] 刷新页面 → 待办仍在
- [ ] 暗色模式 → 整体色板切换,心形仍清晰
- [ ] PWA 装上后,断网再打开 → 列表可看(走 SW 缓存)
- [ ] 后端 `/api/health` 返回 `{"status":"ok","db":"ok"}`

## 11. 风险 & 后续

### 11.1 风险

| 风险 | 缓解 |
|------|------|
| deploy.sh 的"纪念日手动 INSERT"步骤被遗忘 | 部署 README 顶部加 ⚠️ 提示 |
| 心形 + 天数在 iOS 18+ 灵动岛上撞色 | 暗色模式用粉 400 而非 500,实测调 |
| 100 天 milestone 撒花可能惊吓到 | 首次撒花,后续静默(已用 localStorage 记) |
| emoji 在 iOS Safari 显示不一致 | 选 Twemoji stable subset 12 个(🌸🌷🌺🌻🌼🐱🐶🐰🍓☕🍡🎁) |

### 11.2 后续(本设计不做,留口子)

- Web Push 通知(完成后推送)
- 暖话抽屉(预设 10 句)
- 周报(每周日发邮件/推送,总结)
- 待办批量操作(全选/批量勾)
- 待办从菜单菜谱跳转("明天做这道菜" → 加待办)

## 12. 文件改动清单(预演)

**新增后端**:
- `backend/migrations/0005_todo_and_counter.up.sql`
- `backend/migrations/0005_todo_and_counter.down.sql`
- `backend/internal/dao/todo.go`
- `backend/internal/dao/counter.go`
- `backend/internal/api/handlers_todo.go`
- `backend/internal/model/types.go` 追加 `Todo` / `Together` struct

**修改后端**:
- `backend/internal/api/router.go` — 注册 6 个新路由
- `backend/main.go` — 无改(自动跑 0005 migrate)

**新增前端**:
- `frontend/src/pages/TodoPage.vue`
- `frontend/src/components/TabBar.vue`
- `frontend/src/components/TodayTogether.vue`
- `frontend/src/components/TodoInputBar.vue`
- `frontend/src/components/TodoCard.vue`
- `frontend/src/components/DateChipBar.vue`
- `frontend/src/components/EmojiPicker.vue`
- `frontend/src/composables/useTogether.ts`
- `frontend/src/stores/todo.ts`
- `frontend/src/types.ts` 追加 `Todo` / `Together` / `TodoCreate` interface

**修改前端**:
- `frontend/src/router/index.ts` — 加 `/todo` 路由、`/menu/*` namespace
- `frontend/src/App.vue` — 挂 `<TabBar/>`
- `frontend/src/api/client.ts` — 加 6 个方法
- `frontend/src/styles/main.css` — 追加 tab-bar / today-together 相关 token

**不动**:
- `frontend/src/pages/Home.vue`(路由改后变成 `/menu/:week/:date`,内部代码 0 改)
- 所有现有 composables / stores
- `deploy.sh`(除 README 顶部加纪念日提示)

## 13. 决策日志

| 决策点 | 选项 | 选定 | 理由 |
|--------|------|------|------|
| 共享语义 | 个人 / 共享 / 单向留 | 共享 | 双人场景最自然 |
| 完成行为 | 折灰 / 抽屉 / 滑删 | 折灰 | 可撤销 + 沿用 useUndo |
| 组织方式 | 时间分组 / 平铺 / 自定义列表 / 标签 | 平铺 | 轻量 + 翻找时的惊喜感 |
| 截止日期 | 必选 / 可选 / 三类 | 可选 | 灵活 + 兜底"不急" |
| 身份 | 实名 / emoji / 不分 | emoji | 决策成本最低 + 视觉化 |
| tab 栏 | 标准 / 浮起+心 / 抽屉 | 浮起+心 | 让惊喜成为 UI 的一部分 |
| 心形玩法 | 静态+天数 / +暖话 / +每日一句 / +完成夸夸 | 静态+天数 | 克制,先做核心 |
| 后端 | 复用 Go / 拆服务 / 纯前端 | 复用 Go | 共享必须后端,单镜像 |
| 路由 | 单 router / 双 router | 单 router | 简单,共用 `<TabBar/>` |
| TabBar 位置 | 底部 / 顶部 / 抽屉 | 底部 | PWA 移动端规范 |
| 优化更新 | 乐观 / 悲观 | 悲观 | 避免"勾了但 DB 失败" |
| 纪念日存储 | 后端 counter 表 / 前端 localStorage | 后端 | 跨设备一致,跟菜单共用 DB |
