<script setup lang="ts">
import { RouterView } from 'vue-router'
import SiteBrandMark from '@/components/common/SiteBrandMark.vue'
import { useSystemStore } from '@/stores/system'

const system = useSystemStore()
</script>

<template>
  <main class="auth-layout">
    <section class="auth-panel surface">
      <div class="brand">
        <SiteBrandMark :size="48" />
        <div>
          <h1>{{ system.siteName }}</h1>
          <p>多用户在线阅读与图书馆</p>
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

.brand h1 {
  margin: 0;
  font-size: 22px;
}

.brand p {
  margin: 4px 0 0;
  color: var(--color-text-sec);
}
</style>
