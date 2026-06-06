<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
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

const roleOptions = [
  { label: '全部角色', value: null },
  { label: '管理员', value: 'admin' },
  { label: '普通用户', value: 'user' }
]

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' }
]

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
  <PageShell title="用户管理" subtitle="查看用户状态、角色和存储使用">
    <div class="surface filter-bar">
      <n-input v-model:value="filters.keyword" clearable placeholder="搜索用户名或邮箱" @keyup.enter="fetchUsers(1)" />
      <n-select v-model:value="filters.role" :options="roleOptions" />
      <n-select v-model:value="filters.status" :options="statusOptions" />
      <n-button type="primary" secondary @click="fetchUsers(1)">筛选</n-button>
    </div>
    <n-spin :show="loading">
      <div class="surface table">
        <div class="row header">
          <strong>用户</strong>
          <strong>角色</strong>
          <strong>状态</strong>
          <strong>存储</strong>
          <strong>最后登录</strong>
          <strong>操作</strong>
        </div>
        <div v-for="user in users" :key="user.id" class="row">
          <div>
            <strong>{{ user.username }}</strong>
            <span>{{ user.email || '无邮箱' }}</span>
          </div>
          <n-select
            :value="user.role"
            size="small"
            :options="[{ label: '管理员', value: 'admin' }, { label: '普通用户', value: 'user' }]"
            @update:value="(value: UserRole) => onRoleSelect(user, value)"
          />
          <n-select
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
.filter-bar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 150px 150px auto;
  gap: 10px;
  padding: 12px;
}

.table {
  display: grid;
  padding: 12px;
  overflow-x: auto;
}

.row {
  display: grid;
  grid-template-columns: minmax(180px, 2fr) 130px 130px 110px 160px 88px;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--color-border);
}

.row.header {
  color: var(--color-text-sec);
  font-size: 13px;
}

.row div {
  display: grid;
}

.row div span {
  color: var(--color-text-sec);
  font-size: 12px;
}

@media (max-width: 760px) {
  .filter-bar {
    grid-template-columns: 1fr;
  }
}
</style>
