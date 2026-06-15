<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import TabBar from './components/TabBar.vue'
import TodayTogether from './components/TodayTogether.vue'
import UndoToast from './components/UndoToast.vue'
import ThemeToggle from './components/ThemeToggle.vue'

const route = useRoute()
const showHeart = computed(() => route.path.startsWith('/todo'))
</script>

<template>
  <div class="app-shell">
    <RouterView />
    <ThemeToggle />
    <UndoToast />
    <!-- 心形按钮只在 /todo 显示 -->
    <TodayTogether v-if="showHeart" />
    <TabBar />
  </div>
</template>

<style>
#app {
  min-height: 100dvh;
  min-height: 100vh;
}
</style>
