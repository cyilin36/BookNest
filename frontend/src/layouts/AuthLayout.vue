<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useSystemStore } from '@/stores/system'

const system = useSystemStore()
</script>

<template>
  <main class="auth-layout">
    <section class="auth-panel surface">
      <div class="brand">
        <div class="brand-mark">BR</div>
        <div>
          <h1>{{ system.siteName }}</h1>
          <p>多用户在线阅读与公共图书馆</p>
        </div>
      </div>
      <RouterView />
    </section>
  </main>
</template>

<style scoped>
.auth-layout {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 24px;
  background: var(--color-bg-page);
}

.auth-panel {
  width: min(440px, 100%);
  padding: 28px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
}

.brand-mark {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  color: #fff;
  background: var(--color-primary);
  border-radius: 8px;
  font-weight: 800;
}

.brand h1 {
  margin: 0;
  font-size: 22px;
}

.brand p {
  margin: 4px 0 0;
  color: var(--color-text-sec);
}
</style>
