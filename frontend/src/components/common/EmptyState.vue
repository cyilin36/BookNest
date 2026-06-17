<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <div class="empty-state">
    <div class="empty-icon">
      <svg
        width="64"
        height="64"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20" />
      </svg>
    </div>
    <h3 class="empty-title">{{ title }}</h3>
    <p v-if="description" class="empty-description">{{ description }}</p>
    <div v-if="$slots.default" class="empty-actions">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 320px;
  padding: var(--spacing-5xl) var(--spacing-3xl);
  text-align: center;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-large);
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  margin-bottom: var(--spacing-2xl);
  border-radius: var(--radius-round);
  background: var(--color-primary-suppl);
  color: var(--color-primary);
  opacity: 0.8;
}

.empty-icon svg {
  width: 36px;
  height: 36px;
}

.empty-title {
  margin: 0 0 var(--spacing-sm);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-main);
  line-height: var(--line-height-tight);
}

.empty-description {
  margin: 0 0 var(--spacing-2xl);
  font-size: var(--font-size-md);
  color: var(--color-text-sec);
  line-height: var(--line-height-normal);
  max-width: 400px;
}

.empty-actions {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
  justify-content: center;
}

@media (max-width: 768px) {
  .empty-state {
    min-height: 280px;
    padding: var(--spacing-4xl) var(--spacing-2xl);
  }

  .empty-icon {
    width: 64px;
    height: 64px;
    margin-bottom: var(--spacing-lg);
  }

  .empty-icon svg {
    width: 28px;
    height: 28px;
  }

  .empty-title {
    font-size: var(--font-size-xl);
  }

  .empty-description {
    font-size: var(--font-size-base);
  }
}
</style>
