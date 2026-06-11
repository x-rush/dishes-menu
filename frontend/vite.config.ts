import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'auto',
      includeAssets: ['favicon.svg'],
      manifest: {
        name: '每周菜品安排',
        short_name: '菜品',
        description: '给女朋友的私人小项目',
        theme_color: '#f9a8d4',
        background_color: '#fff7f9',
        display: 'standalone',
        orientation: 'portrait',
        start_url: '/',
        scope: '/',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' },
          { src: '/icons/icon-512-maskable.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,svg,png,ico,webmanifest}'],
        navigateFallback: '/index.html',
        runtimeCaching: [
          {
            urlPattern: /^\/api\/dishes/,
            handler: 'NetworkFirst',
            options: { cacheName: 'dishes-api', expiration: { maxAgeSeconds: 60 * 60 } },
          },
          {
            urlPattern: /^\/api\/menu/,
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
