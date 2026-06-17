<script setup lang="ts">
import { RouterView } from 'vue-router'
import TabBar from './components/TabBar.vue'
import UndoToast from './components/UndoToast.vue'
import ThemeToggle from './components/ThemeToggle.vue'
</script>

<template>
  <div class="app-shell">
    <RouterView />
    <ThemeToggle />
    <UndoToast />
    <TabBar />
  </div>
</template>

<style>
#app {
  min-height: 100dvh;
  min-height: 100vh;
}
</style>
