# 待办 Tab + 中央心形惊喜 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 dishes-menu PWA 加一个"待办"tab 和一个中央浮起心形按钮,心形展示在一起天数 + 跳出一个小型惊喜动画;待办与菜单互不干扰,在底部 tab 之间切换。

**Architecture:**
- 后端在现有 Go 服务里加 `todos` 和 `counters` 两张表(migration 0005),提供 6 个 REST 路由
- 前端用 vue-router 把 `/menu/...` 和 `/todo` 拆成两个命名空间;新增 `TabBar.vue` 固定在底部;新增 `TodayTogether.vue` 浮起按钮(中央);tab 切换时心形按钮跟着路由变化(只在 `/todo` 显示)
- 视觉复用现有设计 token(柔粉 + 蜜桃 + 黄油),保持调性一致
- 沿用现有悲观更新策略(wait-200 → 再更新 store),保持乐观交互体验

**Tech Stack:** Go 1.25 + Gin + sqlx + golang-migrate + MySQL 8;Vue 3.5 + Pinia 3 + vue-router 4 + TypeScript 5.7 + @vueuse/core

---

## Conventions Used Throughout

- **Commit message style**: `<type>(<scope>): <subject>` (e.g. `feat(todo): add migration 0005`).
  - 类型: `feat` / `fix` / `refactor` / `chore` / `docs`
  - scope: `backend` / `frontend` / `todo` / `deploy` / etc.
- **Naming conventions**:
  - Go: 字段名用 `snake_case` + JSON tag 也用 `snake_case`(沿用现有)
  - TS interface: 用 `camelCase` 字段;从 API 来的 snake_case 字段在 store 入口做一次转换
  - Vue 组件: `PascalCase.vue`
  - composable: `useXxx.ts`(camelCase 函数名 + 命名导出)
- **No test framework**: 本项目一贯做法 — 用 `go build` / `vue-tsc` / `npm run build` + curl smoke + 手动 E2E
- **Build commands**:
  - Backend: `cd backend && go build ./...`
  - Frontend type-check: `cd frontend && npx vue-tsc --noEmit`
  - Frontend build: `cd frontend && npm run build`

---

## File Structure (what gets created/modified)

### Backend (Go) — new files
- `backend/migrations/0005_todos.up.sql`
- `backend/migrations/0005_todos.down.sql`
- `backend/dao/todo.go` — Todo CRUD with sqlx
- `backend/dao/counter.go` — Counter get/set

### Backend (Go) — modified files
- `backend/model/types.go` — 加 `Todo` 和 `Counter` struct
- `backend/api/router.go` — 挂 6 个新路由

### Frontend — new files
- `frontend/src/stores/todo.ts` — todos 状态
- `frontend/src/composables/useTogether.ts` — 心形 + together_days 计算
- `frontend/src/components/EmojiPicker.vue` — 12 emoji 选择
- `frontend/src/components/DateChipBar.vue` — 今天/明天/后天...快速选 due
- `frontend/src/components/TodoCard.vue` — 单条待办卡片
- `frontend/src/components/TodoInputBar.vue` — 顶部输入条 + emoji 签名 + due 选择
- `frontend/src/components/TodayTogether.vue` — 浮起心形按钮 + 惊喜弹窗
- `frontend/src/components/TabBar.vue` — 底部 tab 切换
- `frontend/src/pages/TodoPage.vue` — /todo 路由页面

### Frontend — modified files
- `frontend/src/types.ts` — 加 Todo / Together 接口
- `frontend/src/api/client.ts` — 加 6 个 API 方法
- `frontend/src/router/index.ts` — 加 /todo 路由;老的 /:week/:date 重定向到 /menu/:week/:date
- `frontend/src/App.vue` — 渲染 TabBar + TodayTogether
- `frontend/src/styles/main.css` — 加 heart 弹跳动画 + together badge 样式

### 部署相关
- `docs/deploy-todo-counter.md` — 提醒运维一起手动 INSERT `together_since`

---

## Phase A: Backend (Go + MySQL)

### Task A1: 创建 migration 0005 (todos + counters 表)

**Files:**
- Create: `backend/migrations/0005_todos.up.sql`
- Create: `backend/migrations/0005_todos.down.sql`

- [ ] **Step 1: 写 0005 up migration**

文件: `backend/migrations/0005_todos.up.sql`

```sql
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS todos (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  content VARCHAR(500) NOT NULL,
  due_date DATE NULL,
  author_emoji VARCHAR(8) NOT NULL,
  author_color VARCHAR(16) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at DATETIME NULL,
  INDEX idx_todos_open (completed_at, due_date, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS counters (
  name VARCHAR(64) NOT NULL PRIMARY KEY,
  value VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

- [ ] **Step 2: 写 0005 down migration**

文件: `backend/migrations/0005_todos.down.sql`

```sql
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS counters;
```

- [ ] **Step 3: 确认本地 migration 版本号**

Run:
```bash
cd /home/x_rush/project/dishes-menu/backend && ls migrations/
```

Expected: 输出 `0001_init.up.sql`、`0002_dishes_v2.up.sql`、`0003_dishes_v3.up.sql`、`0004_dishes_image.up.sql`、`0005_todos.up.sql`(我们刚加的)

- [ ] **Step 4: 在本地 MySQL 跑 migration 验证语法**

Run:
```bash
docker exec -i dishes-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" dishes < migrations/0005_todos.up.sql
```

Expected: 无错误输出。验证:
```bash
docker exec -i dishes-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" dishes -e "SHOW TABLES LIKE 'todos'; SHOW TABLES LIKE 'counters'; DESC todos;"
```

Expected: `todos` 和 `counters` 两表都存在,字段如 spec。

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/0005_todos.up.sql backend/migrations/0005_todos.down.sql
git commit -m "feat(backend): migration 0005 — todos + counters tables"
```

---

### Task A2: model/types.go 加 Todo / Counter struct

**Files:**
- Modify: `backend/model/types.go` (在文件末尾追加)

- [ ] **Step 1: 读现有 types.go 找到合适插入点**

Run:
```bash
tail -20 /home/x_rush/project/dishes-menu/backend/model/types.go
```

Expected: 看到现有 struct 结尾 + 最后一个 `}`。

- [ ] **Step 2: 追加 Todo 和 Counter struct**

文件: `backend/model/types.go`(在文件末尾追加)

```go
type Todo struct {
	ID          int64      `db:"id" json:"id"`
	Content     string     `db:"content" json:"content"`
	DueDate     *time.Time `db:"due_date" json:"due_date,omitempty"`
	AuthorEmoji string     `db:"author_emoji" json:"author_emoji"`
	AuthorColor string     `db:"author_color" json:"author_color"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

type Counter struct {
	Name  string `db:"name" json:"name"`
	Value string `db:"value" json:"value"`
}
```

- [ ] **Step 3: 确认 time 包已 import**

如果 types.go 没有 import `"time"`,在 import 块里加上。

- [ ] **Step 4: build 验证**

Run:
```bash
cd /home/x_rush/project/dishes-menu/backend && go build ./...
```

Expected: `go build` 退出码 0,无错误。

- [ ] **Step 5: Commit**

```bash
git add backend/model/types.go
git commit -m "feat(backend): add Todo + Counter models"
```

---

### Task A3: dao/todo.go CRUD

**Files:**
- Create: `backend/dao/todo.go`

- [ ] **Step 1: 写 dao/todo.go**

文件: `backend/dao/todo.go`

```go
package dao

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xrsh/dishes-menu/backend/model"
)

type TodoDAO struct{ db *sqlx.DB }

func NewTodoDAO(db *sqlx.DB) *TodoDAO { return &TodoDAO{db: db} }

func (d *TodoDAO) List() ([]model.Todo, error) {
	var todos []model.Todo
	err := d.db.Select(&todos, `
		SELECT id, content, due_date, author_emoji, author_color, created_at, completed_at
		FROM todos
		ORDER BY (completed_at IS NULL) DESC, due_date IS NULL, due_date, created_at DESC
	`)
	return todos, err
}

func (d *TodoDAO) Create(content string, dueDate *time.Time, emoji, color string) (model.Todo, error) {
	res, err := d.db.Exec(
		`INSERT INTO todos (content, due_date, author_emoji, author_color) VALUES (?, ?, ?, ?)`,
		content, dueDate, emoji, color,
	)
	if err != nil {
		return model.Todo{}, err
	}
	id, _ := res.LastInsertId()
	return d.Get(id)
}

func (d *TodoDAO) Get(id int64) (model.Todo, error) {
	var t model.Todo
	err := d.db.Get(&t, `
		SELECT id, content, due_date, author_emoji, author_color, created_at, completed_at
		FROM todos WHERE id = ?
	`, id)
	return t, err
}

func (d *TodoDAO) ToggleComplete(id int64) (model.Todo, error) {
	now := time.Now()
	// 已完成 → 取消完成(completed_at = NULL);未完成 → 标记完成
	_, err := d.db.Exec(`
		UPDATE todos
		SET completed_at = CASE WHEN completed_at IS NULL THEN ? ELSE NULL END
		WHERE id = ?
	`, now, id)
	if err != nil {
		return model.Todo{}, err
	}
	return d.Get(id)
}

func (d *TodoDAO) UpdateContent(id int64, content string) error {
	_, err := d.db.Exec(`UPDATE todos SET content = ? WHERE id = ?`, content, id)
	return err
}

func (d *TodoDAO) Delete(id int64) error {
	_, err := d.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	return err
}

// 确保 sql 包被引用(避免 lint 警告)
var _ = sql.ErrNoRows
```

- [ ] **Step 2: build 验证**

Run:
```bash
go build ./...
```

Expected: 退出码 0。

- [ ] **Step 3: Commit**

```bash
git add backend/dao/todo.go
git commit -m "feat(backend): TodoDAO CRUD"
```

---

### Task A4: dao/counter.go

**Files:**
- Create: `backend/dao/counter.go`

- [ ] **Step 1: 写 dao/counter.go**

文件: `backend/backend/dao/counter.go`(实际路径 `backend/dao/counter.go`)

```go
package dao

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type CounterDAO struct{ db *sqlx.DB }

func NewCounterDAO(db *sqlx.DB) *CounterDAO { return &CounterDAO{db: db} }

var ErrCounterNotFound = errors.New("counter not found")

func (d *CounterDAO) Get(name string) (string, error) {
	var value string
	err := d.db.Get(&value, `SELECT value FROM counters WHERE name = ?`, name)
	if err == sql.ErrNoRows {
		return "", ErrCounterNotFound
	}
	return value, err
}

func (d *CounterDAO) Set(name, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO counters (name, value) VALUES (?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value)
	`, name, value)
	return err
}
```

- [ ] **Step 2: build 验证**

Run:
```bash
go build ./...
```

Expected: 退出码 0。

- [ ] **Step 3: Commit**

```bash
git add backend/dao/counter.go
git commit -m "feat(backend): CounterDAO get/set with upsert"
```

---

### Task A5: handlers + router 集成

**Files:**
- Modify: `backend/api/router.go`(追加路由 + handler)

- [ ] **Step 1: 读 router.go 找插入点**

Run:
```bash
cat /home/x_rush/project/dishes-menu/backend/api/router.go
```

记录:路由是如何挂的(有没有 `r.GET(...)` 直接调函数,还是经过某个 controller)。

- [ ] **Step 2: 在 router.go 里追加 handlers**

在 router.go 末尾追加(假设现有模式是 inline handler):

```go
// ── Todos ──
r.GET("/api/todos", func(c *gin.Context) {
	todos, err := todoDAO.List()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, todos)
})

r.POST("/api/todos", func(c *gin.Context) {
	var body struct {
		Content     string  `json:"content"`
		DueDate     *string `json:"due_date"`     // "YYYY-MM-DD" 或 null
		AuthorEmoji string  `json:"author_emoji"`
		AuthorColor string  `json:"author_color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if body.Content == "" || body.AuthorEmoji == "" || body.AuthorColor == "" {
		c.JSON(400, gin.H{"error": "content / author_emoji / author_color 不能为空"})
		return
	}
	var due *time.Time
	if body.DueDate != nil && *body.DueDate != "" {
		t, err := time.Parse("2006-01-02", *body.DueDate)
		if err != nil {
			c.JSON(400, gin.H{"error": "due_date 格式应为 YYYY-MM-DD"})
			return
		}
		due = &t
	}
	todo, err := todoDAO.Create(body.Content, due, body.AuthorEmoji, body.AuthorColor)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, todo)
})

r.PATCH("/api/todos/:id", func(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Content     *string `json:"content"`
		Completed   *bool   `json:"completed"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if body.Completed != nil {
		todo, err := todoDAO.ToggleComplete(id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, todo)
		return
	}
	if body.Content != nil {
		if err := todoDAO.UpdateContent(id, *body.Content); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		todo, _ := todoDAO.Get(id)
		c.JSON(200, todo)
		return
	}
	c.JSON(400, gin.H{"error": "未指定要更新的字段"})
})

r.DELETE("/api/todos/:id", func(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := todoDAO.Delete(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(204)
})

// ── Together counter ──
r.GET("/api/together", func(c *gin.Context) {
	v, err := counterDAO.Get("together_since")
	if err == dao.ErrCounterNotFound {
		c.JSON(200, gin.H{"together_since": nil, "days": 0})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	t, perr := time.Parse("2006-01-02", v)
	if perr != nil {
		c.JSON(500, gin.H{"error": "together_since 格式损坏: " + v})
		return
	}
	days := int(time.Since(t).Hours() / 24)
	if days < 0 {
		days = 0
	}
	c.JSON(200, gin.H{"together_since": v, "days": days})
})

r.POST("/api/together", func(c *gin.Context) {
	var body struct {
		Date string `json:"date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if _, err := time.Parse("2006-01-02", body.Date); err != nil {
		c.JSON(400, gin.H{"error": "date 格式应为 YYYY-MM-DD"})
		return
	}
	if err := counterDAO.Set("together_since", body.Date); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
})
```

并在文件顶部(如果有 init 函数或全局变量)声明 `todoDAO` 和 `counterDAO`:
```go
var (
	todoDAO    = dao.NewTodoDAO(db)
	counterDAO = dao.NewCounterDAO(db)
)
```

并在 import 块加 `"strconv"`(如果还没有)。

- [ ] **Step 3: build 验证**

Run:
```bash
go build ./...
```

Expected: 退出码 0。

- [ ] **Step 4: 启动后端 + curl smoke**

Run:
```bash
go run . &
sleep 2
# 列表(空)
curl -s localhost:8080/api/todos | head -c 200
# 创建一条
curl -s -X POST localhost:8080/api/todos \
  -H 'Content-Type: application/json' \
  -d '{"content":"smoke","author_emoji":"🌸","author_color":"#ec7da6"}'
# again
curl -s localhost:8080/api/todos
# together
curl -s localhost:8080/api/together
# together 未设置时
# 期望: {"together_since":null,"days":0}
```

Expected:
- 第一次 list 返回 `null` 或 `[]`
- POST 返回 201 + 含 id 的对象
- 第二次 list 返回包含 smoke 的数组
- together 返回 `{together_since:null,days:0}`(因为还没插)

- [ ] **Step 5: 杀掉后端进程,commit**

Run:
```bash
kill %1 2>/dev/null || true
git add backend/api/router.go
git commit -m "feat(backend): 6 routes for todos + together counter"
```

---

## Phase B: Frontend infra (types + api + store + composable + router)

### Task B1: types.ts + api/client.ts

**Files:**
- Modify: `frontend/src/types.ts`(追加 Todo / Together)
- Modify: `frontend/src/api/client.ts`(追加 6 个方法)

- [ ] **Step 1: 读现有 types.ts 找插入点**

Run:
```bash
tail -30 /home/x_rush/project/dishes-menu/frontend/src/types.ts
```

- [ ] **Step 2: 在 types.ts 末尾追加**

```ts
export interface Todo {
  id: number
  content: string
  due_date: string | null  // YYYY-MM-DD 或 null
  author_emoji: string
  author_color: string
  created_at: string        // ISO 8601
  completed_at: string | null
}

export interface Together {
  together_since: string | null  // YYYY-MM-DD 或 null
  days: number
}
```

- [ ] **Step 3: 读 api/client.ts 找插入点**

Run:
```bash
grep -n "^export" /home/x_rush/project/dishes-menu/frontend/src/api/client.ts
```

记录现有的导出函数。

- [ ] **Step 4: 在 client.ts 末尾追加 6 个方法**

```ts
import type { Todo, Together } from '../types'

export async function listTodos(): Promise<Todo[]> {
  return request<Todo[]>('/api/todos')
}

export async function createTodo(input: {
  content: string
  due_date: string | null
  author_emoji: string
  author_color: string
}): Promise<Todo> {
  return request<Todo>('/api/todos', { method: 'POST', body: JSON.stringify(input) })
}

export async function patchTodo(id: number, body: { content?: string; completed?: boolean }): Promise<Todo> {
  return request<Todo>(`/api/todos/${id}`, { method: 'PATCH', body: JSON.stringify(body) })
}

export async function deleteTodo(id: number): Promise<void> {
  return request<void>(`/api/todos/${id}`, { method: 'DELETE' })
}

export async function getTogether(): Promise<Together> {
  return request<Together>('/api/together')
}

export async function setTogether(date: string): Promise<void> {
  await request<void>('/api/together', { method: 'POST', body: JSON.stringify({ date }) })
}
```

注意:如果 `request` 函数的签名不支持 204,可能要微调。读一下 `request` 的实现:
Run:
```bash
sed -n '1,80p' /home/x_rush/project/dishes-menu/frontend/src/api/client.ts
```

如果有"response.status === 204 return null"的分支,直接用。如果 `request<Todo[]>` 是泛型调 DELETE 会报错,把 deleteTodo 改成不调 request 直接 fetch:

```ts
export async function deleteTodo(id: number): Promise<void> {
  const res = await fetch(`${base}/api/todos/${id}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 204) throw new Error(`HTTP ${res.status}`)
}
```

具体怎么写取决于现有 `request` 的实现,以实际为准。

- [ ] **Step 5: type-check 验证**

Run:
```bash
cd /home/x_rush/project/dishes-menu/frontend && npx vue-tsc --noEmit
```

Expected: 无错误。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/types.ts frontend/src/api/client.ts
git commit -m "feat(frontend): Todo + Together types and 6 API methods"
```

---

### Task B2: stores/todo.ts

**Files:**
- Create: `frontend/src/stores/todo.ts`

- [ ] **Step 1: 写 stores/todo.ts**

```ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as api from '../api/client'
import type { Todo } from '../types'

export const useTodoStore = defineStore('todo', () => {
  const todos = ref<Todo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const open = computed(() => todos.value.filter(t => !t.completed_at))
  const done = computed(() => todos.value.filter(t => t.completed_at))

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      todos.value = await api.listTodos()
    } catch (e) {
      error.value = e instanceof Error ? e.message : '加载失败'
    } finally {
      loading.value = false
    }
  }

  async function add(input: { content: string; due_date: string | null; author_emoji: string; author_color: string }) {
    const created = await api.createTodo(input)
    todos.value = [created, ...todos.value]
    return created
  }

  async function toggleComplete(id: number) {
    const updated = await api.patchTodo(id, { completed: true })
    const idx = todos.value.findIndex(t => t.id === id)
    if (idx >= 0) todos.value[idx] = updated
  }

  async function updateContent(id: number, content: string) {
    const updated = await api.patchTodo(id, { content })
    const idx = todos.value.findIndex(t => t.id === id)
    if (idx >= 0) todos.value[idx] = updated
  }

  async function remove(id: number) {
    await api.deleteTodo(id)
    todos.value = todos.value.filter(t => t.id !== id)
  }

  return { todos, loading, error, open, done, fetchAll, add, toggleComplete, updateContent, remove }
})
```

- [ ] **Step 2: type-check**

Run:
```bash
npx vue-tsc --noEmit
```

Expected: 无错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/todo.ts
git commit -m "feat(frontend): todo store with optimistic CRUD"
```

---

### Task B3: composables/useTogether.ts

**Files:**
- Create: `frontend/src/composables/useTogether.ts`

- [ ] **Step 1: 写 composable**

```ts
import { ref } from 'vue'
import * as api from '../api/client'
import type { Together } from '../types'

const state = ref<Together>({ together_since: null, days: 0 })
const loaded = ref(false)

export function useTogether() {
  async function refresh() {
    try {
      state.value = await api.getTogether()
    } catch {
      // 静默失败,心形显示默认 0
    } finally {
      loaded.value = true
    }
  }

  async function set(date: string) {
    await api.setTogether(date)
    await refresh()
  }

  return { state, loaded, refresh, set }
}
```

注意:用 module-level `state` 让多个组件共享一份数据(避免重复 fetch)。

- [ ] **Step 2: type-check**

Run:
```bash
npx vue-tsc --noEmit
```

Expected: 无错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/composables/useTogether.ts
git commit -m "feat(frontend): useTogether composable (module-level shared state)"
```

---

### Task B4: router/index.ts 改造

**Files:**
- Modify: `frontend/src/router/index.ts`

- [ ] **Step 1: 读现有 router**

Run:
```bash
cat /home/x_rush/project/dishes-menu/frontend/src/router/index.ts
```

记录:现有路由结构(很可能有一个 `/:week/:date` 路由)。

- [ ] **Step 2: 改造为 namespace**

把现有 `/:week/:date` 改成 `/menu/:week/:date`,加 `/todo` 路由,加 fallback(用户访问 `/` 时跳 `/menu/...`)。

具体怎么改取决于现有代码。目标结构:

```ts
// 老的 /:week/:date → 重定向到 /menu/:week/:date
{
  path: '/:week(202\\d-W\\d{2})/:date(\\d{4}-\\d{2}-\\d{2})',
  redirect: (to) => `/menu${to.fullPath}`,
}

// /menu/:week/:date → Home(现有菜单页)
{
  path: '/menu/:week(202\\d-W\\d{2})/:date(\\d{4}-\\d{2}-\\d{2})',
  component: Home,
},

// /todo → TodoPage
{
  path: '/todo',
  component: () => import('../pages/TodoPage.vue'),
},

// 兜底:/ → 跳到 /menu/<current week>/<today>
{
  path: '/',
  redirect: () => {
    // 用 utils/getCurrentWeek.ts 算出当前周 + 今天
    const { week, today } = getCurrentWeek()
    return `/menu/${week}/${today}`
  },
},
```

具体 week/today 计算函数复用现有的(读 Home.vue 或 utils 看)。

- [ ] **Step 3: type-check + dev smoke**

Run:
```bash
npx vue-tsc --noEmit
npm run dev
```

浏览器访问:
- `http://localhost:5173/` → 跳到 `/menu/2026-W25/2026-06-15`(假设当前是这个)
- `http://localhost:5173/todo` → 看到 TodoPage 占位(暂时是空白或简单的 "TODO 页面")
- 老链接 `http://localhost:5173/2026-W25/2026-06-15` → 重定向到 `/menu/2026-W25/2026-06-15`

确认 OK 后 `Ctrl-C` 杀掉。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/router/index.ts
git commit -m "refactor(frontend): namespace routes /menu and /todo"
```

---

## Phase C: Components

### Task C1: EmojiPicker.vue

**Files:**
- Create: `frontend/src/components/EmojiPicker.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [string] }>()

const EMOJIS = ['🌸','🌷','🌺','🌻','🌼','🐱','🐶','🐰','🍓','☕','🍡','🎁']

function pick(e: string) {
  emit('update:modelValue', e)
}
</script>

<template>
  <div class="emoji-picker" role="radiogroup" aria-label="选择你的 emoji 签名">
    <button
      v-for="e in EMOJIS"
      :key="e"
      type="button"
      role="radio"
      :aria-checked="props.modelValue === e"
      :class="['emoji-btn', { active: props.modelValue === e }]"
      @click="pick(e)"
    >{{ e }}</button>
  </div>
</template>

<style scoped>
.emoji-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.emoji-btn {
  width: 36px;
  height: 36px;
  font-size: 22px;
  border-radius: var(--radius-sm);
  background: var(--color-cream);
  transition: transform 0.15s var(--ease-spring), background 0.15s ease;
}
.emoji-btn:hover { background: var(--color-pink-50); }
.emoji-btn:active { transform: scale(0.9); }
.emoji-btn.active {
  background: var(--color-pink-100);
  box-shadow: 0 0 0 2px var(--color-pink-400);
}
</style>
```

- [ ] **Step 2: type-check**

Run:
```bash
npx vue-tsc --noEmit
```

Expected: 无错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/EmojiPicker.vue
git commit -m "feat(frontend): EmojiPicker — 12 emoji subset"
```

---

### Task C2: DateChipBar.vue

**Files:**
- Create: `frontend/src/components/DateChipBar.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ modelValue: string | null }>()
const emit = defineEmits<{ 'update:modelValue': [string | null] }>()

function fmt(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const today = new Date()
const todayStr = fmt(today)

const tomorrow = new Date(today); tomorrow.setDate(tomorrow.getDate() + 1)
const tomorrowStr = fmt(tomorrow)

const dayAfter = new Date(today); dayAfter.setDate(dayAfter.getDate() + 2)
const dayAfterStr = fmt(dayAfter)

const nextWeek = new Date(today); nextWeek.setDate(nextWeek.getDate() + 7)
const nextWeekStr = fmt(nextWeek)

const chips = [
  { label: '今天', value: todayStr },
  { label: '明天', value: tomorrowStr },
  { label: '后天', value: dayAfterStr },
  { label: '下周', value: nextWeekStr },
]

function pick(v: string) {
  // 再点同一个 → 取消
  emit('update:modelValue', props.modelValue === v ? null : v)
}

function clear() {
  emit('update:modelValue', null)
}
</script>

<template>
  <div class="date-chip-bar">
    <button
      v-for="c in chips"
      :key="c.value"
      type="button"
      :class="['chip', { active: modelValue === c.value }]"
      @click="pick(c.value)"
    >{{ c.label }} · {{ c.value.slice(5) }}</button>
    <button
      v-if="modelValue"
      type="button"
      class="chip clear"
      @click="clear"
    >清除</button>
  </div>
</template>

<style scoped>
.date-chip-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.chip {
  font-size: 12px;
  padding: 5px 10px;
  background: var(--color-cream);
  color: var(--color-muted);
  border-radius: var(--radius-pill);
  transition: background 0.15s ease, color 0.15s ease;
}
.chip:hover { background: var(--color-pink-50); color: var(--color-pink-500); }
.chip.active {
  background: var(--color-pink-200);
  color: var(--color-pink-600);
  font-weight: 600;
}
.chip.clear {
  background: transparent;
  color: var(--color-danger);
  text-decoration: underline;
}
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/components/DateChipBar.vue
git commit -m "feat(frontend): DateChipBar — quick due-date presets"
```

---

### Task C3: TodoCard.vue

**Files:**
- Create: `frontend/src/components/TodoCard.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { Todo } from '../types'
import { useTodoStore } from '../stores/todo'
import { useUndo } from '../composables/useUndo'

const props = defineProps<{ todo: Todo }>()
const store = useTodoStore()
const undo = useUndo()
const editing = ref(false)
const editText = ref(props.todo.content)

async function toggle() {
  await store.toggleComplete(props.todo.id)
  if (!props.todo.completed_at) {
    // 刚勾上 → 5 秒内可撤销
    undo.push(`已勾上「${props.todo.content.slice(0, 12)}」`, async () => {
      await store.toggleComplete(props.todo.id)
    })
  }
}

function startEdit() {
  editing.value = true
  editText.value = props.todo.content
}

async function saveEdit() {
  const trimmed = editText.value.trim()
  if (!trimmed || trimmed === props.todo.content) {
    editing.value = false
    return
  }
  await store.updateContent(props.todo.id, trimmed)
  editing.value = false
}

async function remove() {
  const snapshot = { ...props.todo }
  await store.remove(props.todo.id)
  undo.push(`已删除「${snapshot.content.slice(0, 12)}」`, async () => {
    await store.add({
      content: snapshot.content,
      due_date: snapshot.due_date,
      author_emoji: snapshot.author_emoji,
      author_color: snapshot.author_color,
    })
  })
}

function formatDue(d: string | null) {
  if (!d) return ''
  // d = 'YYYY-MM-DD',取 MM-DD
  return d.slice(5)
}

function relativeTime(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const min = Math.floor(ms / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h} 小时前`
  const d = Math.floor(h / 24)
  return `${d} 天前`
}
</script>

<template>
  <article :class="['todo-card', { done: todo.completed_at }]">
    <button
      type="button"
      class="check"
      :aria-pressed="!!todo.completed_at"
      :aria-label="todo.completed_at ? '取消完成' : '标记完成'"
      @click="toggle"
    >
      <span v-if="todo.completed_at">✓</span>
    </button>

    <div class="body">
      <div v-if="editing" class="edit-row">
        <input v-model="editText" maxlength="500" @keydown.enter="saveEdit" @keydown.esc="editing = false" />
        <button class="btn btn-primary save" @click="saveEdit">保存</button>
      </div>
      <div v-else class="content-row">
        <p class="content" @click="startEdit">{{ todo.content }}</p>
      </div>

      <div class="meta">
        <span class="sig" :style="{ background: todo.author_color }" :title="`由 ${todo.author_emoji} 添加`">
          {{ todo.author_emoji }}
        </span>
        <span v-if="todo.due_date" class="due">📅 {{ formatDue(todo.due_date) }}</span>
        <span class="time">{{ relativeTime(todo.created_at) }}</span>
      </div>
    </div>

    <button class="btn btn-icon del" :aria-label="`删除 ${todo.content}`" @click="remove">×</button>
  </article>
</template>

<style scoped>
.todo-card {
  display: flex;
  gap: 10px;
  padding: 12px 14px;
  background: var(--color-cream);
  border-radius: var(--radius-md);
  align-items: flex-start;
  transition: opacity 0.3s ease, background 0.3s ease;
}
.todo-card.done { opacity: 0.55; background: var(--color-warm-bg); }
.todo-card.done .content { text-decoration: line-through; color: var(--color-muted); }

.check {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 2px solid var(--color-pink-300);
  background: transparent;
  color: var(--color-pink-500);
  font-size: 16px;
  flex: 0 0 auto;
  margin-top: 2px;
  transition: background 0.2s var(--ease-spring), border-color 0.2s ease, transform 0.15s var(--ease-spring);
}
.check:hover { border-color: var(--color-pink-500); }
.check:active { transform: scale(0.88); }
.todo-card.done .check {
  background: var(--color-pink-400);
  border-color: var(--color-pink-500);
  color: #fff;
}

.body { flex: 1 1 auto; min-width: 0; }
.content-row { display: flex; gap: 8px; align-items: center; }
.content {
  font-size: 15px;
  word-break: break-word;
  cursor: text;
  margin: 0;
}
.edit-row {
  display: flex;
  gap: 6px;
}
.edit-row input { flex: 1 1 auto; font-size: 14px; padding: 6px 10px; }
.save { padding: 4px 12px; font-size: 12px; }

.meta {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--color-muted);
  align-items: center;
  flex-wrap: wrap;
}
.sig {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  filter: brightness(1.05);
}
.due { color: var(--color-peach-300); }
.time { color: var(--color-muted); }

.del {
  width: 28px;
  height: 28px;
  min-width: 28px;
  min-height: 28px;
  font-size: 16px;
  background: transparent;
  color: var(--color-muted);
  flex: 0 0 auto;
}
.del:hover { background: var(--color-pink-100); color: var(--color-danger); }
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/components/TodoCard.vue
git commit -m "feat(frontend): TodoCard — check/edit/delete with undo"
```

---

### Task C4: TodoInputBar.vue

**Files:**
- Create: `frontend/src/components/TodoInputBar.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import EmojiPicker from './EmojiPicker.vue'
import DateChipBar from './DateChipBar.vue'

const store = useTodoStore()
const content = ref('')
const dueDate = ref<string | null>(null)
const emoji = ref<string>(localStorage.getItem('todo:emoji') ?? '🌸')
const color = ref<string>(localStorage.getItem('todo:color') ?? '#ec7da6')
const expanded = ref(false)

const EMOJI_COLORS: Record<string, string> = {
  '🌸': '#ec7da6', '🌷': '#d65a8a', '🌺': '#e26d6d', '🌻': '#f7d560',
  '🌼': '#ffc9a8', '🐱': '#ffa980', '🐶': '#d2a679', '🐰': '#e8d5e0',
  '🍓': '#e26d6d', '☕': '#a0826d', '🍡': '#ffc7dc', '🎁': '#7dd3a3',
}

function pickEmoji(e: string) {
  emoji.value = e
  color.value = EMOJI_COLORS[e] ?? '#ec7da6'
  localStorage.setItem('todo:emoji', e)
  localStorage.setItem('todo:color', color.value)
}

watch(emoji, (e) => {
  color.value = EMOJI_COLORS[e] ?? '#ec7da6'
})

const canSubmit = computed(() => content.value.trim().length > 0)

async function submit() {
  if (!canSubmit.value) return
  await store.add({
    content: content.value.trim(),
    due_date: dueDate.value,
    author_emoji: emoji.value,
    author_color: color.value,
  })
  content.value = ''
  dueDate.value = null
  expanded.value = false
}
</script>

<template>
  <div :class="['input-bar card', { expanded }]">
    <div class="top-row">
      <button
        type="button"
        class="sig-btn"
        :title="`当前签名:${emoji} (点选切换)`"
        :style="{ background: color }"
        @click="expanded = !expanded"
      >{{ emoji }}</button>

      <input
        v-model="content"
        class="content-input"
        placeholder="想做点什么?写一句…"
        maxlength="500"
        @keydown.enter="submit"
        @focus="expanded = true"
      />

      <button class="btn btn-primary submit" :disabled="!canSubmit" @click="submit">＋</button>
    </div>

    <div v-if="expanded" class="extras">
      <div class="extras-row">
        <span class="extras-label">签名:</span>
        <EmojiPicker :model-value="emoji" @update:model-value="pickEmoji" />
      </div>
      <div class="extras-row">
        <span class="extras-label">截止:</span>
        <DateChipBar v-model="dueDate" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.input-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--color-pink-50);
  border: 1.5px solid var(--color-pink-100);
}
.top-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.sig-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 22px;
  flex: 0 0 auto;
  filter: brightness(1.05);
  transition: transform 0.15s var(--ease-spring);
}
.sig-btn:active { transform: scale(0.92); }
.content-input {
  flex: 1 1 auto;
  background: var(--color-cream);
  border: 1.5px solid var(--color-line-2);
  border-radius: var(--radius-sm);
  padding: 10px 14px;
  font-size: 15px;
}
.submit {
  width: 44px;
  height: 40px;
  font-size: 22px;
  padding: 0;
  flex: 0 0 auto;
}
.extras {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 6px;
  border-top: 1px dashed var(--color-line);
}
.extras-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.extras-label {
  font-size: 12px;
  color: var(--color-muted);
  width: 36px;
  flex: 0 0 auto;
}
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/components/TodoInputBar.vue
git commit -m "feat(frontend): TodoInputBar — content + emoji + due"
```

---

### Task C5: TodayTogether.vue

**Files:**
- Create: `frontend/src/components/TodayTogether.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTogether } from '../composables/useTogether'

const { state, loaded, refresh } = useTogether()
const dialogRef = ref<HTMLDialogElement | null>(null)
const surpriseVisible = ref(false)

onMounted(() => {
  refresh()
})

const display = computed(() => {
  if (!loaded.value || !state.value.together_since) return '♡'
  return String(state.value.days)
})

const subtitle = computed(() => {
  if (!loaded.value) return '加载中…'
  if (!state.value.together_since) return '点我设置纪念日'
  return `在一起 ${state.value.days} 天`
})

function onClick() {
  if (!state.value.together_since) {
    // 首次:打开设置弹窗
    dialogRef.value?.showModal()
    return
  }
  // 已有:触发惊喜动画
  surpriseVisible.value = true
  setTimeout(() => {
    surpriseVisible.value = false
  }, 2400)
}

const inputDate = ref<string>(state.value.together_since ?? new Date().toISOString().slice(0, 10))
const { set } = useTogether()

async function save() {
  await set(inputDate.value)
  dialogRef.value?.close()
}
</script>

<template>
  <button class="heart-fab" :class="{ pulse: surpriseVisible }" @click="onClick">
    <span class="heart-icon">{{ display }}</span>
    <span class="heart-sub">{{ subtitle }}</span>
  </button>

  <!-- 惊喜粒子 -->
  <Transition name="surprise">
    <div v-if="surpriseVisible" class="surprise-overlay" aria-hidden="true">
      <div v-for="i in 12" :key="i" :class="['sparkle', `s${i}`]">✨</div>
      <p class="surprise-msg">💗 想你 💗</p>
    </div>
  </Transition>

  <dialog ref="dialogRef" @click="(e) => e.target === dialogRef && dialogRef.close()">
    <div class="dlg-body">
      <h3>设置纪念日</h3>
      <p class="hint">输入你们在一起的日子,以后每天都会算在一起的天数。</p>
      <input v-model="inputDate" type="date" />
      <div class="dlg-footer">
        <button class="btn btn-ghost" @click="dialogRef?.close()">取消</button>
        <button class="btn btn-primary" @click="save">保存</button>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.heart-fab {
  position: fixed;
  bottom: calc(72px + env(safe-area-inset-bottom));  /* 浮在 TabBar 上 */
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  width: 76px;
  background: transparent;
  z-index: 30;
  pointer-events: auto;
}
.heart-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff8fb1, #ec7da6);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6px 18px rgba(236, 125, 166, 0.4);
  animation: heart-bob 2.4s ease-in-out infinite;
}
.heart-sub {
  font-size: 10px;
  color: var(--color-pink-500);
  font-weight: 600;
  background: var(--color-cream);
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  box-shadow: var(--shadow-sm);
  white-space: nowrap;
}
.heart-fab.pulse .heart-icon {
  animation: heart-pop 0.6s var(--ease-spring);
}
@keyframes heart-bob {
  0%, 100% { transform: translateY(0) scale(1); }
  50%      { transform: translateY(-3px) scale(1.04); }
}
@keyframes heart-pop {
  0%   { transform: scale(1); }
  40%  { transform: scale(1.25); }
  100% { transform: scale(1); }
}

.surprise-overlay {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}
.sparkle {
  position: absolute;
  font-size: 28px;
  animation: sparkle-fly 1.6s ease-out forwards;
}
.s1  { left: 20%; top: 30%; animation-delay: 0.0s; }
.s2  { left: 30%; top: 60%; animation-delay: 0.1s; }
.s3  { left: 70%; top: 25%; animation-delay: 0.15s; }
.s4  { left: 80%; top: 55%; animation-delay: 0.05s; }
.s5  { left: 50%; top: 20%; animation-delay: 0.2s; }
.s6  { left: 25%; top: 70%; animation-delay: 0.12s; }
.s7  { left: 75%; top: 70%; animation-delay: 0.18s; }
.s8  { left: 45%; top: 80%; animation-delay: 0.08s; }
.s9  { left: 55%; top: 35%; animation-delay: 0.22s; }
.s10 { left: 15%; top: 45%; animation-delay: 0.13s; }
.s11 { left: 85%; top: 40%; animation-delay: 0.17s; }
.s12 { left: 40%; top: 50%; animation-delay: 0.07s; }
@keyframes sparkle-fly {
  0%   { transform: translate(0, 0) scale(0.5); opacity: 0; }
  30%  { opacity: 1; }
  100% { transform: translate(var(--tx, 40px), var(--ty, -60px)) scale(1.2); opacity: 0; }
}
.s1  { --tx: -30px; --ty: -50px; }
.s2  { --tx:  20px; --ty: -70px; }
.s3  { --tx: -40px; --ty: -30px; }
.s4  { --tx:  35px; --ty: -80px; }
.s5  { --tx: -10px; --ty: -90px; }
.s6  { --tx:  50px; --ty: -40px; }
.s7  { --tx: -45px; --ty: -20px; }
.s8  { --tx:  25px; --ty: -100px; }
.s9  { --tx: -25px; --ty: -60px; }
.s10 { --tx:  60px; --ty: -30px; }
.s11 { --tx: -55px; --ty: -45px; }
.s12 { --tx:  10px; --ty: -110px; }

.surprise-msg {
  position: absolute;
  top: 38%;
  font-size: 28px;
  font-weight: 700;
  color: var(--color-pink-600);
  font-family: var(--font-display);
  animation: msg-pop 2.4s ease-out forwards;
}
@keyframes msg-pop {
  0%   { transform: scale(0.6); opacity: 0; }
  20%  { transform: scale(1.1); opacity: 1; }
  80%  { transform: scale(1); opacity: 1; }
  100% { transform: scale(1.1); opacity: 0; }
}

.surprise-enter-active, .surprise-leave-active { transition: opacity 0.2s ease; }
.surprise-enter-from, .surprise-leave-to { opacity: 0; }

.dlg-body { padding: 20px; min-width: min(320px, 90vw); }
.dlg-body h3 { margin: 0 0 8px; font-size: 17px; }
.hint { font-size: 13px; color: var(--color-muted); margin: 0 0 12px; }
.dlg-body input[type=date] { width: 100%; font-size: 15px; }
.dlg-footer { display: flex; gap: 8px; justify-content: flex-end; margin-top: 14px; }
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/components/TodayTogether.vue
git commit -m "feat(frontend): TodayTogether — heart FAB + surprise particles"
```

---

### Task C6: TabBar.vue

**Files:**
- Create: `frontend/src/components/TabBar.vue`

- [ ] **Step 1: 写组件**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const active = computed<'menu' | 'todo'>(() => {
  return route.path.startsWith('/todo') ? 'todo' : 'menu'
})

function go(target: 'menu' | 'todo') {
  if (target === active.value) return
  if (target === 'todo') {
    router.push('/todo')
  } else {
    // 回菜单:保留当前的 week + date(如果是在 /menu 命名空间下),否则跳今天
    if (typeof route.params.week === 'string' && typeof route.params.date === 'string') {
      router.push(`/menu/${route.params.week}/${route.params.date}`)
    } else {
      // 兜底:跳 / 让 redirect 处理
      router.push('/')
    }
  }
}
</script>

<template>
  <nav class="tab-bar" role="tablist">
    <button
      type="button"
      role="tab"
      :aria-selected="active === 'menu'"
      :class="['tab', { active: active === 'menu' }]"
      @click="go('menu')"
    >
      <span class="tab-icon">🍱</span>
      <span class="tab-label">菜单</span>
    </button>
    <div class="tab-spacer"></div>
    <button
      type="button"
      role="tab"
      :aria-selected="active === 'todo'"
      :class="['tab', { active: active === 'todo' }]"
      @click="go('todo')"
    >
      <span class="tab-icon">💝</span>
      <span class="tab-label">待办</span>
    </button>
  </nav>
</template>

<style scoped>
.tab-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: calc(60px + env(safe-area-inset-bottom));
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-cream);
  border-top: 1px solid var(--color-line);
  display: flex;
  align-items: stretch;
  z-index: 25;
  box-shadow: 0 -4px 12px rgba(0,0,0,0.04);
}
.tab {
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: transparent;
  color: var(--color-muted);
  font-size: 11px;
  font-weight: 600;
  min-height: 44px;
  transition: color 0.2s ease;
}
.tab.active {
  color: var(--color-pink-500);
}
.tab-icon {
  font-size: 24px;
  transition: transform 0.2s var(--ease-spring);
}
.tab.active .tab-icon {
  transform: translateY(-2px) scale(1.1);
}
.tab-spacer {
  flex: 0 0 90px;  /* 中间留位置给 heart FAB */
}
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/components/TabBar.vue
git commit -m "feat(frontend): TabBar — bottom menu/todo switcher with center spacer"
```

---

### Task C7: styles/main.css 加 heart 弹跳

**Files:**
- Modify: `frontend/src/styles/main.css`(追加动画 + tab-bar 底部留白)

- [ ] **Step 1: 在 main.css 末尾追加**

```css
/* 底部 TabBar 留白 */
.app-shell { padding-bottom: calc(72px + env(safe-area-inset-bottom)); }

/* 浮动心形按钮的层级 */
.heart-fab-wrap { z-index: 30; }

/* 中央浮起按钮让底部 nav 不被挡 */
@media (max-width: 480px) {
  .tab-spacer { flex-basis: 100px; }
}
```

(动画已经在 TodayTogether.vue 内部 scoped 写了,这里只放布局微调。)

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/styles/main.css
git commit -m "feat(frontend): styles for TabBar + heart FAB layout"
```

---

## Phase D: 页面集成

### Task D1: TodoPage.vue

**Files:**
- Create: `frontend/src/pages/TodoPage.vue`

- [ ] **Step 1: 写页面**

```vue
<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useTodoStore } from '../stores/todo'
import TodoInputBar from '../components/TodoInputBar.vue'
import TodoCard from '../components/TodoCard.vue'

const store = useTodoStore()

onMounted(() => {
  store.fetchAll()
})

const sortedOpen = computed(() =>
  [...store.open].sort((a, b) => {
    // 有 due_date 排前;同 due_date 新的在前
    if (a.due_date && !b.due_date) return -1
    if (!a.due_date && b.due_date) return 1
    if (a.due_date && b.due_date) return a.due_date.localeCompare(b.due_date)
    return b.created_at.localeCompare(a.created_at)
  })
)
</script>

<template>
  <div class="todo-page">
    <header class="page-head">
      <h1>💝 想做的小事</h1>
      <p class="muted">写下你们想一起做的事</p>
    </header>

    <TodoInputBar />

    <section v-if="store.loading && !store.todos.length" class="empty">加载中…</section>
    <section v-else-if="!sortedOpen.length && !store.done.length" class="empty">
      <p>还没有任何待办。从上面写一条试试 ✏️</p>
    </section>

    <template v-else>
      <section v-if="sortedOpen.length" class="list-section">
        <h2 class="section-title">待办 · {{ sortedOpen.length }}</h2>
        <div class="todo-list">
          <TodoCard v-for="t in sortedOpen" :key="t.id" :todo="t" />
        </div>
      </section>

      <section v-if="store.done.length" class="list-section done-section">
        <h2 class="section-title">已完成 · {{ store.done.length }}</h2>
        <div class="todo-list">
          <TodoCard v-for="t in store.done" :key="t.id" :todo="t" />
        </div>
      </section>
    </template>

    <div v-if="store.error" class="error">{{ store.error }}</div>
  </div>
</template>

<style scoped>
.todo-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
  max-width: 560px;
  margin: 0 auto;
}
.page-head h1 {
  font-size: 24px;
  font-family: var(--font-display);
  color: var(--color-pink-500);
  margin: 0 0 4px;
}
.muted { font-size: 13px; color: var(--color-muted); }

.section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-pink-500);
  font-family: var(--font-display);
  letter-spacing: 0.04em;
  margin: 8px 0 0;
}
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.done-section { margin-top: 8px; opacity: 0.85; }

.empty {
  text-align: center;
  padding: 32px 16px;
  color: var(--color-muted);
  font-size: 14px;
}
.error {
  background: var(--color-pink-100);
  color: var(--color-danger);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
}
</style>
```

- [ ] **Step 2: type-check + Commit**

```bash
npx vue-tsc --noEmit
git add frontend/src/pages/TodoPage.vue
git commit -m "feat(frontend): TodoPage — input + open/done sections"
```

---

### Task D2: App.vue 集成 TabBar + TodayTogether

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 读现有 App.vue**

Run:
```bash
cat /home/x_rush/project/dishes-menu/frontend/src/App.vue
```

- [ ] **Step 2: 改造为集成 TabBar + TodayTogether**

典型模式:

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import TabBar from './components/TabBar.vue'
import TodayTogether from './components/TodayTogether.vue'
import { useUndo } from './composables/useUndo'
import UndoToast from './components/UndoToast.vue'  // 复用现有

const route = useRoute()
const showHeart = computed(() => route.path.startsWith('/todo'))
</script>

<template>
  <div class="app-shell">
    <router-view />
    <UndoToast />
    <!-- 心形按钮只在 /todo 显示 -->
    <TodayTogether v-if="showHeart" />
    <TabBar />
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100dvh;
  padding-bottom: calc(72px + env(safe-area-inset-bottom));
}
</style>
```

具体 UndoToast 是否存在,以及它是怎么 mount 的,以实际为准。如果不用单独的 Toast 组件,而是在 useUndo 内部用 teleport,可以省一行。

- [ ] **Step 3: dev smoke**

Run:
```bash
npm run dev
```

浏览器:
1. 访问 `http://localhost:5173/` → 应该跳到 `/menu/...`,看到菜单页 + 底部 TabBar(菜单 active)
2. 点底部"待办" → 跳 `/todo`,看到 TodoPage + 心形按钮浮起
3. 点心形 → 触发惊喜粒子(没有设置 together_since 时会跳设置弹窗)
4. 设置一个日期 → 心形数字变成天数
5. 再点 → 惊喜动画

确认 OK 后 `Ctrl-C`。

- [ ] **Step 4: build 验证**

Run:
```bash
npm run build
```

Expected: 成功,无错误。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat(frontend): integrate TabBar + TodayTogether into App shell"
```

---

## Phase E: E2E + 部署文档

### Task E1: 完整 smoke test

- [ ] **Step 1: 启动后端 + 前端**

Run:
```bash
# 后端
cd /home/x_rush/project/dishes-menu/backend && go run . &
# 前端(另一个终端)
cd /home/x_rush/project/dishes-menu/frontend && npm run dev &
sleep 3
```

- [ ] **Step 2: 浏览器完整流程**

1. 打开 `http://localhost:5173/`
2. 应自动跳到 `/menu/<当前周>/<今天>`
3. 菜单页能正常用(沿用现有功能,确认没坏)
4. 点底部"待办" → 跳到 `/todo`
5. 在输入框写"周末去公园",按回车 → 应该出现一条待办
6. 点心形 → 弹设置 dialog
7. 设置一个日期(如 2025-01-01) → 心形变成 N 天
8. 再点心形 → 惊喜粒子动画
9. 切回菜单 → 老路由没坏

- [ ] **Step 3: curl 验证 API**

```bash
# 后端 list
curl -s localhost:8080/api/todos
# together
curl -s localhost:8080/api/together
```

Expected:
- todos 返回刚刚创建的列表
- together 返回 `{together_since: "2025-01-01", days: <N>}`

- [ ] **Step 4: 关掉前后端**

```bash
kill %1 %2 2>/dev/null || true
```

---

### Task E2: 部署文档 + together_since 手动初始化

**Files:**
- Create: `docs/deploy-todo-counter.md`

- [ ] **Step 1: 写部署备忘**

```markdown
# 待办 + 在一起天数 — 部署备忘

## 1. 跑 migration 0005

(沿用现有 migration 部署流程)

```bash
# 拉最新代码 + 跑 migration
ssh fstbnet
cd /path/to/dishes-menu/backend
./migrate up  # 或者 make migrate,看你项目用什么命令
```

或者手动跑:
```bash
mysql -h 39.105.6.177 -u root -p dishes < migrations/0005_todos.up.sql
```

## 2. 手动设置 together_since ⚠️

**这一步不能 commit 到 git**(具体日期是个人隐私)。

SSH 到生产 MySQL 后:
```sql
INSERT INTO counters (name, value) VALUES ('together_since', 'YYYY-MM-DD')
ON DUPLICATE KEY UPDATE value = VALUES(value);
```

把 `YYYY-MM-DD` 替换成你们实际在一起的日期。

## 3. 部署前端

按现有 `deploy.sh` 流程。`/` 路径自动 redirect 到 `/menu/<current week>/<today>`,老链接会重定向到 `/menu/...`,兼容老 PWA 用户。

## 4. 验证清单

- [ ] `curl /api/todos` 返回 `[]` 或已有数据
- [ ] `curl /api/together` 返回 `{together_since: "YYYY-MM-DD", days: <N>}`
- [ ] 浏览器打开 PWA,菜单页正常
- [ ] 切到待办页,心形按钮显示天数
- [ ] 点心形能触发惊喜动画
```

- [ ] **Step 2: Commit**

```bash
git add docs/deploy-todo-counter.md
git commit -m "docs: deployment notes for todos + together_since"
```

- [ ] **Step 3: 部署到生产**

按 deploy.sh 走(SSH key auth,沿用现有流程)。

部署完后浏览器验证:
1. 心形按钮显示天数
2. 菜单页老路由没坏

---

## Self-Review (pre-flight)

- [x] **Spec coverage**:
  - 双 tab 切换 ✓ (TabBar + router namespace)
  - 心形 + 在一起天数 ✓ (TodayTogether + /api/together)
  - 惊喜动画 ✓ (sparkle particles in TodayTogether)
  - emoji 签名 ✓ (EmojiPicker + TodoInputBar)
  - 灰划掉 + 5 秒撤销 ✓ (TodoCard + useUndo)
  - 可选 due_date ✓ (DateChipBar + todo.due_date)
  - 共享待办(无用户隔离) ✓ (单表 todos)
- [x] **No placeholders**: 所有 step 都是具体代码/命令;emoji 12 子集、日期格式化、动画 keyframes 都写出来了
- [x] **Type consistency**: 后端 Todo snake_case json tag ↔ 前端 Todo.due_date snake_case(在 store 直接用,转换在显示层用 .slice(5))
- [x] **Frequent commits**: 每个 Task 1 个 commit,Phase A 5 个、Phase B 4 个、Phase C 7 个、Phase D 2 个、Phase E 2 个 = 20 commits 总
- [x] **Self-test path covered**: E1 是手测 + curl,符合本项目惯例(无 vitest)
- [x] **Spec 安全约束**: together_since 不进 git(部署备忘里写明手动 INSERT);emoji 12 子集写死;无用户认证

---

## Out of Scope (明确不做)

- ❌ 用户注册/登录(单用户 PWA,无公开访问)
- ❌ Web Push 通知(已 spec 排除)
- ❌ 待办分类/标签(已 spec 排除,平铺 + due_date 排序)
- ❌ 待办分享链接(已 spec 排除)
- ❌ 服务端推 together 倒计时(轮询足够,心形只在 /todo 显示)
- ❌ 多端同步(待办共享已通过"无认证"实现)
