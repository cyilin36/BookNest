<script setup lang="ts">
import { computed, ref, h } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { BookOpen, ChartNoAxesColumn, ChevronDown, Library, LogOut, Menu, Settings, Shield, Tags, Upload, Users, User } from 'lucide-vue-next'
import SiteBrandMark from '@/components/common/SiteBrandMark.vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const system = useSystemStore()
const mobileOpen = ref(false)
const userMenuOpen = ref(false)

const navItems = computed(() => [
  { path: '/bookshelf', label: '我的书架', icon: BookOpen },
  { path: '/library', label: '图书馆', icon: Library },
  { path: '/upload', label: '上传', icon: Upload },
  { path: '/settings', label: '设置', icon: Settings }
])

const adminItems = [
  { path: '/admin', label: '概览', icon: ChartNoAxesColumn },
  { path: '/admin/users', label: '用户管理', icon: Users },
  { path: '/admin/library', label: '图书管理', icon: Library },
  { path: '/admin/categories', label: '分类管理', icon: Tags },
  { path: '/admin/tags', label: '标签管理', icon: Tags },
  { path: '/admin/settings', label: '系统设置', icon: Settings }
]

function closeMobile() {
  mobileOpen.value = false
}

function closeUserMenu() {
  userMenuOpen.value = false
}

async function logout() {
  await auth.logout()
  router.push('/login')
}

const adminDropdownOptions = computed(() =>
  adminItems.map((item) => ({
    label: item.label,
    key: item.path
  }))
)

const userDropdownOptions = computed(() => [
  {
    label: auth.user?.role === 'admin' ? '管理员' : '普通用户',
    key: 'role',
    disabled: true
  },
  {
    type: 'divider',
    key: 'd1'
  },
  {
    label: '个人资料',
    key: 'profile',
    icon: () => h(User, { size: 18 })
  },
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(LogOut, { size: 18 })
  }
])

function handleAdminSelect(key: string) {
  router.push(key)
}

function handleUserSelect(key: string) {
  if (key === 'logout') {
    logout()
  } else if (key === 'profile') {
    router.push('/profile')
  }
}
</script>

<template>
  <n-layout class="app-layout">
    <!-- Desktop Top Navigation -->
    <header class="top-navbar">
      <div class="navbar-container">
        <div class="navbar-left">
          <SiteBrandMark :size="32" />
          <strong class="site-name">{{ system.siteName }}</strong>
        </div>

        <nav class="navbar-center">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-tab"
            :class="{ active: route.path.startsWith(item.path) }"
          >
            {{ item.label }}
          </RouterLink>
          <n-dropdown v-if="auth.isAdmin" :options="adminDropdownOptions" @select="handleAdminSelect">
            <button class="nav-tab" :class="{ active: route.path.startsWith('/admin') }">
              管理
              <ChevronDown :size="16" />
            </button>
          </n-dropdown>
        </nav>

        <div class="navbar-right">
          <n-dropdown :options="userDropdownOptions" @select="handleUserSelect">
            <button class="user-button">
              <div v-if="auth.user?.avatar_url" class="user-avatar">
                <img :src="`${auth.user.avatar_url}?t=${Date.now()}`" alt="用户头像" class="avatar-image" />
              </div>
              <div v-else class="user-avatar">
                <User :size="20" />
              </div>
              <span class="user-name">{{ auth.user?.nickname || auth.user?.username }}</span>
            </button>
          </n-dropdown>
        </div>

        <!-- Mobile hamburger -->
        <button class="mobile-menu-btn" @click="mobileOpen = true">
          <Menu :size="24" />
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <n-layout-content class="content">
      <RouterView />
    </n-layout-content>
  </n-layout>

  <!-- Mobile Drawer -->
  <n-drawer v-model:show="mobileOpen" placement="left" :width="280">
    <n-drawer-content :title="system.siteName">
      <nav class="mobile-nav-list">
        <RouterLink
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="mobile-nav-item"
          :class="{ active: route.path.startsWith(item.path) }"
          @click="closeMobile"
        >
          <component :is="item.icon" :size="20" />
          <span>{{ item.label }}</span>
        </RouterLink>

        <template v-if="auth.isAdmin">
          <div class="mobile-nav-divider">管理功能</div>
          <RouterLink
            v-for="item in adminItems"
            :key="item.path"
            :to="item.path"
            class="mobile-nav-item"
            :class="{ active: route.path === item.path || (item.path === '/admin/library' && route.path === '/admin/books') }"
            @click="closeMobile"
          >
            <component :is="item.icon" :size="20" />
            <span>{{ item.label }}</span>
          </RouterLink>
        </template>
      </nav>
      <template #footer>
        <n-button block secondary @click="logout">
          <template #icon>
            <LogOut :size="18" />
          </template>
          退出登录
        </n-button>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: var(--color-bg-page);
}

/* Top Navigation Bar */
.top-navbar {
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border);
}

.navbar-container {
  max-width: var(--container-max-width);
  margin: 0 auto;
  display: flex;
  align-items: center;
  height: var(--height-navbar);
  padding: 0 var(--spacing-3xl);
  gap: var(--spacing-5xl);
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.site-name {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-main);
  letter-spacing: var(--letter-spacing-tight);
}

.navbar-center {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex: 1;
}

.nav-tab {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  height: 36px;
  padding: var(--spacing-sm) var(--spacing-lg);
  color: var(--color-text-sec);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-small);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all var(--transition-base);
  text-decoration: none;
  white-space: nowrap;
}

.nav-tab:hover {
  color: var(--color-text-main);
  background: var(--color-bg-page);
}

.nav-tab.active {
  color: #ffffff;
  background: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.user-button {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  height: 36px;
  padding: var(--spacing-xs) var(--spacing-md) var(--spacing-xs) var(--spacing-xs);
  background: transparent;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  transition: all var(--transition-base);
  color: var(--color-text-main);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
}

.user-button:hover {
  background: var(--color-bg-hover);
}

.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-round);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  overflow: hidden;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-menu-btn {
  display: none;
  align-items: center;
  justify-content: center;
  width: var(--height-button-md);
  height: var(--height-button-md);
  background: transparent;
  border: none;
  border-radius: var(--radius-small);
  cursor: pointer;
  color: var(--color-text-main);
  transition: background var(--transition-base);
}

.mobile-menu-btn:hover {
  background: var(--color-bg-hover);
}

/* Main Content */
.content {
  max-width: var(--container-max-width);
  margin: 0 auto;
  padding: 0;
  min-height: calc(100vh - var(--height-navbar));
}

/* Mobile Navigation */
.mobile-nav-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.mobile-nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  height: 48px;
  padding: 0 var(--spacing-lg);
  color: var(--color-text-sec);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-small);
  transition: all var(--transition-base);
  text-decoration: none;
}

.mobile-nav-item:hover,
.mobile-nav-item.active {
  color: var(--color-primary);
  background: var(--color-primary-suppl);
}

.mobile-nav-divider {
  padding: var(--spacing-lg) var(--spacing-lg) var(--spacing-sm);
  color: var(--color-text-sec);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: var(--letter-spacing-widest);
}

/* Responsive */
@media (max-width: 1024px) {
  .navbar-center {
    gap: var(--spacing-xs);
  }

  .nav-tab {
    padding: var(--spacing-sm) var(--spacing-md);
    font-size: var(--font-size-base);
  }
}

@media (max-width: 768px) {
  .navbar-container {
    height: var(--height-navbar-mobile);
    padding: 0 var(--padding-page-mobile);
    gap: var(--spacing-lg);
  }

  .navbar-center {
    display: none;
  }

  .navbar-right {
    display: none;
  }

  .mobile-menu-btn {
    display: flex;
    margin-left: auto;
  }

  .content {
    min-height: calc(100vh - var(--height-navbar-mobile));
  }
}

@media (max-width: 520px) {
  .site-name {
    display: none;
  }
}
</style>
