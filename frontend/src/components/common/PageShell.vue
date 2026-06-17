<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <section class="page-shell">
    <div class="container">
      <header class="page-header">
        <div class="page-header-content">
          <h1 class="page-title">{{ title }}</h1>
          <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
        </div>
        <div v-if="$slots.actions" class="page-actions">
          <slot name="actions" />
        </div>
      </header>
      <div class="page-content">
        <slot />
      </div>
    </div>
  </section>
</template>

<style scoped>
.page-shell {
  min-height: calc(100vh - var(--height-navbar));
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-3xl);
  margin-bottom: var(--spacing-section);
}

.page-header-content {
  flex: 1;
  min-width: 0;
}

.page-title {
  margin: 0;
  font-size: var(--font-size-5xl);
  font-weight: var(--font-weight-bold);
  line-height: var(--line-height-tight);
  letter-spacing: var(--letter-spacing-tighter);
  color: var(--color-text-main);
}

.page-subtitle {
  margin: var(--spacing-sm) 0 0;
  font-size: var(--font-size-md);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
}

.page-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.page-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-section);
}

/* Responsive */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-lg);
  }

  .page-title {
    font-size: var(--font-size-3xl);
  }

  .page-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 520px) {
  .page-title {
    font-size: var(--font-size-3xl);
  }

  .page-subtitle {
    font-size: var(--font-size-base);
  }
}
</style>
