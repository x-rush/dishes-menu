package api

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"dishes-menu/internal/dao"
	"dishes-menu/internal/service"
)

type Handlers struct {
	Health *HealthHandler
	Dish   *DishHandler
	Menu   *MenuHandler
}

func NewHandlers(db *sqlx.DB) *Handlers {
	dishRepo := dao.NewDishRepo(db)
	menuRepo := dao.NewMenuRepo(db)
	shuffle := service.NewShuffleService(dishRepo, menuRepo)
	return &Handlers{
		Health: NewHealthHandler(db),
		Dish:   NewDishHandler(dishRepo),
		Menu:   NewMenuHandler(menuRepo, dishRepo, shuffle),
	}
}

func RegisterRoutes(r *gin.Engine, h *Handlers, webFS fs.FS) {
	r.Use(corsMiddleware())
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/", "/favicon.ico", "/manifest.webmanifest", "/sw.js"},
	}))

	api := r.Group("/api")
	{
		api.GET("/health", h.Health.Health)
		api.GET("/dishes", h.Dish.List)
		api.POST("/dishes", h.Dish.Create)
		api.PUT("/dishes/:id", h.Dish.Update)
		api.DELETE("/dishes/:id", h.Dish.Delete)
		api.GET("/menu", h.Menu.GetWeek)
		api.POST("/menu/:day/:slot", h.Menu.AppendItem)
		api.PUT("/menu/:day/:slot/:seq", h.Menu.UpdateItemNote)
		api.DELETE("/menu/:day/:slot/:seq", h.Menu.DeleteItem)
		api.GET("/menu/shuffle", h.Menu.Shuffle)
	}

	if webFS != nil {
		serveStatic(r, webFS)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func serveStatic(r *gin.Engine, webFS fs.FS) {
	fileServer := http.FileServer(http.FS(webFS))
	r.GET("/", func(c *gin.Context) { serveIndex(c, webFS) })
	r.GET("/index.html", func(c *gin.Context) { serveIndex(c, webFS) })
	r.NoRoute(func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if p == "" {
			serveIndex(c, webFS)
			return
		}
		// Vite base = /forxt/dishes-menu/ 时,index.html 引用的资源路径是
		// /forxt/dishes-menu/assets/...。但 webFS 嵌入的是 dist/ root(没有
		// sub-path 前缀)。生产 nginx 会把 /forxt/dishes-menu/ strip 后再 proxy,
		// 所以后端看到的是 /assets/...;但裸 :8080(本地 dev / 容器直连)看到
		// 的是完整 /forxt/dishes-menu/assets/...。两种都剥掉 sub-path 前缀后
		// 再查 webFS,命中就 serve。
		fsPath := strings.TrimPrefix(p, "forxt/dishes-menu/")
		if fsPath == p {
			// 没匹配上 sub-path prefix,直接用 p
			fsPath = p
		}
		if tryServeAsset(c, fileServer, webFS, fsPath) {
			return
		}
		serveIndex(c, webFS)
	})
}

// tryServeAsset 在 webFS 中查 filePath,命中且带扩展名就 serve;否则返回 false 让 caller 走 SPA fallback。
func tryServeAsset(c *gin.Context, fileServer http.Handler, webFS fs.FS, filePath string) bool {
	if _, err := fs.Stat(webFS, filePath); err == nil {
		if hasExt(filePath) {
			c.Header("Content-Type", mime.TypeByExtension(path.Ext(filePath)))
			c.Request.URL.Path = "/" + filePath
			fileServer.ServeHTTP(c.Writer, c.Request)
			return true
		}
	}
	return false
}

func serveIndex(c *gin.Context, webFS fs.FS) {
	data, err := fs.ReadFile(webFS, "index.html")
	if err != nil {
		c.String(http.StatusNotFound, "frontend not built; run `npm run build` in frontend/")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func hasExt(p string) bool {
	dot := strings.LastIndex(p, ".")
	slash := strings.LastIndex(p, "/")
	return dot > slash && dot != -1
}
