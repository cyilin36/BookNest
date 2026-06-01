<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { BookOpen, Library, LogOut, Menu, Settings, Shield, Upload } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const system = useSystemStore()
const mobileOpen = ref(false)

const navItems = computed(() => [
  { path: '/bookshelf', label: '我的书架', icon: BookOpen },
  { path: '/library', label: '公共馆', icon: Library },
  { path: '/upload', label: '上传', icon: Upload },
  { path: '/settings', label: '设置', icon: Settings },
  ...(auth.isAdmin ? [{ path: '/admin', label: '管理', icon: Shield }] : [])
])

async function logout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <n-layout class="app-layout" has-sider>
    <n-layout-sider class="sidebar" :width="236" bordered collapse-mode="width">
      <div class="sidebar-brand">
        <div class="brand-mark">BR</div>
        <strong>{{ system.siteName }}</strong>
      </div>
      <nav class="nav-list">
        <RouterLink v-for="item in navItems" :key="item.path" :to="item.path" class="nav-item" :class="{ active: route.path.startsWith(item.path) }">
          <component :is="item.icon" :size="18" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <div class="sidebar-user">
        <div>
          <strong>{{ auth.user?.nickname || auth.user?.username }}</strong>
          <span>{{ auth.user?.role === 'admin' ? '管理员' : '普通用户' }}</span>
        </div>
        <n-button quaternary circle title="退出登录" @click="logout">
          <LogOut :size="18" />
        </n-button>
      </div>
    </n-layout-sider>
    <n-layout>
      <header class="mobile-topbar">
        <n-button quaternary circle @click="mobileOpen = true"><Menu :size="20" /></n-button>
        <strong>{{ system.siteName }}</strong>
      </header>
      <n-layout-content class="content">
        <RouterView />
      </n-layout-content>
    </n-layout>
  </n-layout>
  <n-drawer v-model:show="mobileOpen" placement="left" :width="280">
    <n-drawer-content :title="system.siteName">
      <nav class="nav-list mobile">
        <RouterLink v-for="item in navItems" :key="item.path" :to="item.path" class="nav-item" @click="mobileOpen = false">
          <component :is="item.icon" :size="18" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <template #footer>
        <n-button block @click="logout">退出登录</n-button>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.sidebar {
  background: var(--color-bg-sidebar);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 64px;
  padding: 0 18px;
}

.brand-mark {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  color: #fff;
  background: var(--color-primary);
  border-radius: 8px;
  font-size: 12px;
  font-weight: 800;
}

.nav-list {
  display: grid;
  gap: 4px;
  padding: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 42px;
  padding: 0 12px;
  color: var(--color-text-sec);
  border-radius: 8px;
}

.nav-item.active,
.nav-item:hover {
  color: var(--color-primary);
  background: var(--color-primary-suppl);
}

.sidebar-user {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 14px 16px;
  border-top: 1px solid var(--color-border);
}

.sidebar-user div {
  display: grid;
  min-width: 0;
}

.sidebar-user span {
  color: var(--color-text-sec);
  font-size: 12px;
}

.content {
  min-height: 100vh;
  padding: 28px;
  background: var(--color-bg-page);
}

.mobile-topbar {
  display: none;
  align-items: center;
  gap: 12px;
  height: 56px;
  padding: 0 14px;
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border);
}

@media (max-width: 900px) {
  .sidebar {
    display: none;
  }

  .mobile-topbar {
    display: flex;
  }

  .content {
    padding: 16px;
  }
}
</style>
