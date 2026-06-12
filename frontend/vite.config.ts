import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  base: '/forxt/dishes-menu/',
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
        start_url: '/forxt/dishes-menu/',
        scope: '/forxt/dishes-menu/',
        icons: [
          { src: '/forxt/dishes-menu/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/forxt/dishes-menu/icons/icon-512.png', sizes: '512x512', type: 'image/png' },
          { src: '/forxt/dishes-menu/icons/icon-512-maskable.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,png,ico,webmanifest,woff2}'],
        navigateFallback: '/forxt/dishes-menu/index.html',
        runtimeCaching: [
          {
            urlPattern: /^\/forxt\/dishes-menu\/api\/dishes/,
            handler: 'NetworkFirst',
            options: { cacheName: 'dishes-api', expiration: { maxAgeSeconds: 60 * 60 } },
          },
          {
            urlPattern: /^\/forxt\/dishes-menu\/api\/menu/,
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
  build: {
    outDir: '../backend/web/dist',
    emptyOutDir: true,
    target: 'es2020',
  },
})
