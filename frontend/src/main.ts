import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/fonts.css'
import './styles/main.css'
import { registerSW } from 'virtual:pwa-register'

// 动态修正 <link rel="icon"> href:dev 直接 /favicon.svg,prod 拼 sub-path 前缀
const favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
if (favicon) {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  favicon.href = base + '/favicon.svg'
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')

if ('serviceWorker' in navigator && import.meta.env.PROD) {
  registerSW({ immediate: true })
}
