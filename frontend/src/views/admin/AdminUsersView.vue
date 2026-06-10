<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { Search, Trash2 } from 'lucide-vue-next'
import PageShell from '@/components/common/PageShell.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { userApi } from '@/api/user'
import type { Pagination, User, UserRole, UserStatus } from '@/api/types'
import { formatBytes, formatDate } from '@/utils/format'

const message = useMessage()
const dialog = useDialog()
const users = ref<User[]>([])
const loading = ref(false)
const deletingUserId = ref<number | null>(null)
const pagination = ref<Pagination>({ page: 1, page_size: 20, total: 0 })
const filters = reactive({
  keyword: '',
  role: null as UserRole | null,
  status: null as UserStatus | null
})

const roleOptions: { label: string; value: UserRole | null }[] = [
  { label: '全部角色', value: null },
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' }
]

const statusOptions: { label: string; value: UserStatus | null }[] = [
  { label: '全部状态', value: null },
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' }
]

const activeUsersOnPage = computed(() => users.value.filter((user) => user.status === 'active').length)
const loadedRange = computed(() => {
  if (!users.value.length) return '0'
  const start = (pagination.value.page - 1) * pagination.value.page_size + 1
  const end = start + users.value.length - 1
  return `${start}-${end}`
})

async function fetchUsers(page = 1) {
  loading.value = true
  try {
    const result = await userApi.list({
      page,
      page_size: pagination.value.page_size,
      keyword: filters.keyword || undefined,
      role: filters.role || undefined,
      status: filters.status || undefined
    })
    users.value = result.data
    pagination.value = result.pagination
  } finally {
    loading.value = false
  }
}

function setRoleFilter(role: UserRole | null) {
  filters.role = role
}

function setStatusFilter(status: UserStatus | null) {
  filters.status = status
}

async function changeRole(user: User, role: UserRole) {
  try {
    Object.assign(user, await userApi.updateRole(user.id, role))
    message.success('角色已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新失败')
  }
}

async function changeStatus(user: User, status: UserStatus) {
  try {
    Object.assign(user, await userApi.updateStatus(user.id, status))
    message.success('状态已更新')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新失败')
  }
}

function confirmDeleteUser(user: User) {
  dialog.error({
    title: '删除用户账号',
    content: `确定删除普通用户「${user.username}」吗？该操作会清空该账号、图书、书架、阅读进度、书签等所有个人数据，删除后无法恢复。`,
    positiveText: '删除账号',
    negativeText: '取消',
    onPositiveClick: async () => {
      deletingUserId.value = user.id
      try {
        await userApi.remove(user.id)
        message.success('用户已删除')
        await fetchUsers(pagination.value.page)
      } catch (error) {
        message.error(error instanceof Error ? error.message : '删除失败')
      } finally {
        deletingUserId.value = null
      }
    }
  })
}

function onRoleSelect(user: User, value: UserRole) {
  changeRole(user, value)
}

function onStatusSelect(user: User, value: UserStatus) {
  changeStatus(user, value)
}

onMounted(() => fetchUsers())
</script>

<template>
  <PageShell class="admin-users-page" title="用户管理" subtitle="查看用户状态、角色和存储使用">
    <template #actions>
      <div class="user-stats">
        <span>总用户 {{ pagination.total }}</span>
        <span>本页启用 {{ activeUsersOnPage }}</span>
      </div>
    </template>
    <section class="surface user-filter-panel">
      <label class="filter-field search-field">
        <span>搜索</span>
        <div class="filter-control">
          <n-input v-model:value="filters.keyword" clearable placeholder="搜索用户名或邮箱" @keyup.enter="fetchUsers(1)">
            <template #prefix><Search :size="16" /></template>
          </n-input>
        </div>
      </label>
      <div class="filter-field">
        <span>角色</span>
        <div class="segmented-control">
          <button
            v-for="option in roleOptions"
            :key="String(option.value)"
            type="button"
            :class="{ active: filters.role === option.value }"
            @click="setRoleFilter(option.value)"
          >
            {{ option.label.replace('全部', '全部') }}
          </button>
        </div>
      </div>
      <div class="filter-field">
        <span>状态</span>
        <div class="segmented-control">
          <button
            v-for="option in statusOptions"
            :key="String(option.value)"
            type="button"
            :class="{ active: filters.status === option.value }"
            @click="setStatusFilter(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>
      <n-button class="filter-submit" type="primary" @click="fetchUsers(1)">
        <template #icon><Search :size="16" /></template>
        搜索
      </n-button>
    </section>
    <n-spin :show="loading">
      <div class="surface users-table-card">
        <div class="table-meta">
          <strong>用户列表</strong>
          <span>显示 {{ loadedRange }} / 共 {{ pagination.total }}</span>
        </div>
        <div class="row header">
          <strong>用户资料</strong>
          <strong>角色</strong>
          <strong>状态</strong>
          <strong>存储</strong>
          <strong>最后登录</strong>
          <strong>操作</strong>
        </div>
        <div v-for="user in users" :key="user.id" class="row">
          <div class="user-profile">
            <span class="avatar">{{ user.username.slice(0, 1).toUpperCase() }}</span>
            <span>
              <strong>{{ user.username }}</strong>
              <small>{{ user.email || '无邮箱' }}</small>
            </span>
          </div>
          <n-select
            class="cell-select"
            :value="user.role"
            size="small"
            :options="[{ label: '管理员', value: 'admin' }, { label: '普通用户', value: 'user' }]"
            @update:value="(value: UserRole) => onRoleSelect(user, value)"
          />
          <n-select
            class="cell-select"
            :value="user.status"
            size="small"
            :options="[{ label: '启用', value: 'active' }, { label: '禁用', value: 'disabled' }]"
            @update:value="(value: UserStatus) => onStatusSelect(user, value)"
          />
          <span>{{ formatBytes(user.storage_used_bytes) }}</span>
          <span>{{ formatDate(user.last_login_at) }}</span>
          <n-button
            v-if="user.role === 'user'"
            size="small"
            type="error"
            secondary
            :loading="deletingUserId === user.id"
            @click="confirmDeleteUser(user)"
          >
            <template #icon><Trash2 :size="15" /></template>
            删除
          </n-button>
          <span v-else class="muted">-</span>
        </div>
      </div>
    </n-spin>
    <PaginationBar :pagination="pagination" @change="fetchUsers" />
  </PageShell>
</template>

<style scoped>
.admin-users-page {
  gap: 16px;
}

.user-stats {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
  color: var(--color-text-sec);
  font-size: 13px;
}

.user-stats span {
  padding: 6px 10px;
  background: var(--color-primary-suppl);
  border-radius: 999px;
  color: var(--color-primary);
  font-weight: 700;
}

.user-filter-panel {
  display: grid;
  grid-template-columns: minmax(220px, 1.25fr) minmax(220px, auto) minmax(180px, auto) auto;
  align-items: end;
  gap: 14px;
  padding: 14px;
}

.filter-field {
  display: grid;
  min-width: 0;
  gap: 7px;
}

.filter-field > span {
  color: var(--color-text-main);
  font-size: 13px;
  font-weight: 800;
}

.filter-control {
  min-width: 0;
}

.segmented-control {
  display: flex;
  min-height: 34px;
  padding: 3px;
  background: #eef2f5;
  border-radius: 8px;
}

.segmented-control button {
  min-width: 0;
  padding: 0 10px;
  color: var(--color-text-sec);
  background: transparent;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 700;
  white-space: nowrap;
}

.segmented-control button.active {
  color: var(--color-text-main);
  background: #ffffff;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.12);
}

.filter-submit {
  min-width: 82px;
}

.users-table-card {
  display: grid;
  padding: 0 14px 10px;
  overflow-x: auto;
}

.table-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 760px;
  padding: 14px 4px 10px;
}

.table-meta span {
  color: var(--color-text-sec);
  font-size: 13px;
}

.row {
  display: grid;
  grid-template-columns: minmax(220px, 2fr) 132px 132px 110px 160px 88px;
  align-items: center;
  gap: 12px;
  min-width: 760px;
  padding: 11px 4px;
  border-bottom: 1px solid var(--color-border);
}

.row.header {
  color: var(--color-text-sec);
  font-size: 13px;
}

.user-profile {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 10px;
}

.user-profile > span:last-child {
  display: grid;
  min-width: 0;
}

.user-profile strong,
.user-profile small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-profile small {
  color: var(--color-text-sec);
  font-size: 12px;
}

.avatar {
  display: inline-grid;
  flex: 0 0 34px;
  width: 34px;
  height: 34px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-suppl);
  border-radius: 999px;
  font-size: 14px;
  font-weight: 900;
}

.cell-select {
  max-width: 132px;
}

@media (max-width: 1080px) {
  .user-filter-panel {
    grid-template-columns: 1fr 1fr;
  }

  .search-field {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .admin-users-page :deep(.page-header) {
    flex-direction: column;
  }

  .user-stats {
    justify-content: flex-start;
  }

  .user-filter-panel {
    grid-template-columns: 1fr;
  }

  .segmented-control {
    overflow-x: auto;
  }
}
</style>
