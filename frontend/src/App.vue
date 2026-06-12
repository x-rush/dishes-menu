<script setup lang="ts">
import { RouterView } from 'vue-router'
import UndoToast from './components/UndoToast.vue'
import ThemeToggle from './components/ThemeToggle.vue'
</script>

<template>
  <RouterView />
  <ThemeToggle />
  <UndoToast />
</template>

<style>
#app {
  min-height: 100dvh;
  min-height: 100vh;
}
</style>
