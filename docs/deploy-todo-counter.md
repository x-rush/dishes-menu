# 待办 + 在一起天数 — 部署备忘

> 配套 [实施计划](../superpowers/plans/2026-06-15-todo-tab-surprise.md) 的 Phase E 部署文档。

## 1. 跑 migration 0005

0005 包含两张新表:`todos`(待办)和 `counters`(键值对,目前只存 `together_since`)。

**走 deploy.sh**:build 阶段会自动跑迁移,无需手动操作。

**手动跑**(如果只更新 DB 不重部署后端):

```bash
mysql -h 39.105.6.177 -u root -p dishes_menu < backend/migrations/0005_todos.up.sql
```

回滚:

```bash
mysql -h 39.105.6.177 -u root -p dishes_menu < backend/migrations/0005_todos.down.sql
```

migration 文件首行是 `SET NAMES utf8mb4;`,确保中文 content 字段不会出现双重编码。

## 2. 手动设置 together_since ⚠️

**这一步永远不要 commit 到 git**(具体日期是个人隐私,跟谁无关)。

SSH 到生产服务器后:

```bash
mysql -h 39.105.6.177 -u root -p dishes_menu
```

```sql
INSERT INTO counters (name, value) VALUES ('together_since', 'YYYY-MM-DD')
ON DUPLICATE KEY UPDATE value = VALUES(value);
```

把 `YYYY-MM-DD` 替换成你们实际在一起的日期。`ON DUPLICATE KEY UPDATE` 让改日期也能用同一语句。

`days` 字段由后端实时算 `int(time.Since(t).Hours()/24)`,前端只读不写,不需要手动维护。

## 3. 部署前端

按现有 `deploy.sh` 流程:

```bash
cd /home/x_rush/project/dishes-menu
./deploy.sh
```

构建产物通过 `//go:embed` 编进 Go 二进制,最终镜像 ~15 MB distroless/static。

## 4. 路由变化

- `/` 自动 redirect 到 `/menu/<current-week>/<today>`,老 PWA 链接不会 404。
- `/menu/:week/:date` 仍是菜单主页。
- 新增 `/todo` 路由。
- 老路由(无 namespace 的)被 router 的 redirect 规则兜住。

## 5. 验证清单

部署完后逐项打勾:

- [ ] `curl https://<host>/api/todos` 返回 `[]` 或已有数据
- [ ] `curl https://<host>/api/together` 返回 `{"together_since":"YYYY-MM-DD","days":<N>}`
- [ ] 浏览器打开 PWA,菜单页正常
- [ ] 切到底部"待办"tab,看到 💝 心形 FAB
- [ ] 心形显示正确天数(如果步骤 2 已设置)
- [ ] 点心形触发惊喜粒子动画
- [ ] 输入框输入"周末去公园"按回车,出现一条待办
- [ ] 心形旁边的勾可以划掉待办,5 秒内可撤销

## 6. 回滚

如果出问题:

```bash
# 回滚后端到上一版本
cd /mnt/dockerv/dishes-menu && docker compose down
docker compose up -d  # 用上一版本镜像

# 单独回滚 migration(慎用,会丢失 todos 数据)
mysql -h 39.105.6.177 -u root -p dishes_menu < backend/migrations/0005_todos.down.sql
```

前端是 SPA,新版本兼容旧版 DB schema(老用户没数据时不会崩)。
