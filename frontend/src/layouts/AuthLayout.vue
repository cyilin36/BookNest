<script setup lang="ts">
import { computed } from 'vue'
import { RouterView } from 'vue-router'
import SiteBrandMark from '@/components/common/SiteBrandMark.vue'
import { useSystemStore } from '@/stores/system'

const system = useSystemStore()

const backgroundStyle = computed(() => {
  if (system.loginBackgroundSrc) {
    return {
      backgroundImage: `url(${system.loginBackgroundSrc})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat'
    }
  }
  return {
    background: 'linear-gradient(135deg, #E0F2FE 0%, #BAE6FD 100%)'
  }
})
</script>

<template>
  <main class="auth-layout" :style="backgroundStyle">
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
  padding: var(--spacing-3xl);
}

.auth-panel {
  width: min(480px, 100%);
  padding: var(--spacing-4xl);
  background: var(--color-bg-card);
  border-radius: var(--radius-xlarge);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.08);
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-3xl);
  padding-bottom: var(--spacing-2xl);
  border-bottom: 1px solid var(--color-border);
}

.brand h1 {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-main);
  letter-spacing: -0.02em;
}

.brand p {
  margin: var(--spacing-xs) 0 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-sec);
  line-height: var(--line-height-tight);
}

@media (max-width: 520px) {
  .auth-layout {
    padding: var(--spacing-xl);
  }

  .auth-panel {
    padding: var(--spacing-2xl);
  }

  .brand {
    gap: var(--spacing-md);
    margin-bottom: var(--spacing-2xl);
  }

  .brand h1 {
    font-size: var(--font-size-xl);
  }
}
</style>
