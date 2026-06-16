import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

// 双模式 base path:本地 dev = /,生产 = /forxt/dishes-menu/(走 .env 文件切换)
// 注意:process.env.VITE_* 在 vite.config.ts 求值时是 undefined(env 文件只注入到
// import.meta.env,不会出现在 config 阶段的 process.env),必须用 loadEnv 显式读。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const BASE = env.VITE_BASE_PATH || '/'
  const BASE_NORMALIZED = BASE.endsWith('/') ? BASE : BASE + '/'
  return {
    base: BASE,
    plugins: [
      vue(),
      VitePWA({
        registerType: 'autoUpdate',
        injectRegister: 'auto',
        includeAssets: ['favicon.svg'],
        manifest: {
          name: '今天想吃什么呀',
          short_name: '好好吃饭',
          description: '好可爱&她的餐桌',
          theme_color: '#f9a8d4',
          background_color: '#fff7f9',
          display: 'standalone',
          orientation: 'portrait',
          start_url: BASE,
          scope: BASE,
          icons: [
            { src: `${BASE_NORMALIZED}icons/icon-192.png`, sizes: '192x192', type: 'image/png' },
            { src: `${BASE_NORMALIZED}icons/icon-512.png`, sizes: '512x512', type: 'image/png' },
            { src: `${BASE_NORMALIZED}icons/icon-512-maskable.png`, sizes: '512x512', type: 'image/png', purpose: 'maskable' },
          ],
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,svg,png,ico,webmanifest,woff2}'],
          navigateFallback: `${BASE}index.html`,
          runtimeCaching: [
            {
              urlPattern: new RegExp(`^${BASE.replace(/\//g, '\\/')}api\\/dishes`),
              handler: 'NetworkFirst',
              options: { cacheName: 'dishes-api', expiration: { maxAgeSeconds: 60 * 60 } },
            },
            {
              urlPattern: new RegExp(`^${BASE.replace(/\//g, '\\/')}api\\/menu`),
              handler: 'NetworkFirst',
              options: { cacheName: 'menu-api', expiration: { maxAgeSeconds: 60 } },
            },
          ],
        },
        devOptions: { enabled: false },
      }),
    ],
    server: {
      port: 5173,
      proxy: { '/api': { target: 'http://localhost:8080', changeOrigin: true } },
    },
    css: {
      postcss: {
        plugins: [
          {
            postcssPlugin: 'replace-base-url',
            Once(root) {
              root.walkAtRules('font-face', (rule) => {
                rule.walkDecls('src', (decl) => {
                  if (typeof decl.value === 'string') {
                    decl.value = decl.value.replace(/%BASE_URL%/g, BASE_NORMALIZED)
                  }
                })
              })
            },
          },
        ],
      },
    },
    build: {
      outDir: '../backend/web/dist',
      emptyOutDir: true,
      target: 'es2020',
    },
  }
})
