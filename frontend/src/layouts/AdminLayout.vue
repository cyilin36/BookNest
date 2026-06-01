<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { ChartNoAxesColumn, Library, Settings, Tags, Users } from 'lucide-vue-next'

const route = useRoute()
const links = [
  { path: '/admin', label: '概览', icon: ChartNoAxesColumn },
  { path: '/admin/users', label: '用户', icon: Users },
  { path: '/admin/library', label: '公共图书', icon: Library },
  { path: '/admin/categories', label: '分类', icon: Tags },
  { path: '/admin/tags', label: '标签', icon: Tags },
  { path: '/admin/settings', label: '系统', icon: Settings }
]
</script>

<template>
  <main class="admin-layout">
    <aside class="admin-nav surface">
      <RouterLink to="/bookshelf" class="back-link">返回书架</RouterLink>
      <RouterLink v-for="link in links" :key="link.path" :to="link.path" class="admin-link" :class="{ active: route.path === link.path }">
        <component :is="link.icon" :size="17" />
        <span>{{ link.label }}</span>
      </RouterLink>
    </aside>
    <section class="admin-content">
      <RouterView />
    </section>
  </main>
</template>

<style scoped>
.admin-layout {
  display: grid;
  grid-template-columns: 210px minmax(0, 1fr);
  min-height: 100vh;
  gap: 18px;
  padding: 20px;
  background: var(--color-bg-page);
}

.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-self: start;
  padding: 12px;
  position: sticky;
  top: 20px;
}

.back-link,
.admin-link {
  display: flex;
  align-items: center;
  gap: 9px;
  min-height: 40px;
  padding: 0 10px;
  color: var(--color-text-sec);
  border-radius: 8px;
}

.back-link {
  margin-bottom: 8px;
  color: var(--color-primary);
}

.admin-link.active,
.admin-link:hover {
  color: var(--color-primary);
  background: var(--color-primary-suppl);
}

@media (max-width: 840px) {
  .admin-layout {
    grid-template-columns: 1fr;
    padding: 14px;
  }

  .admin-nav {
    position: static;
    flex-flow: row wrap;
  }
}
</style>
