<script setup lang="ts">
import { reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()
const form = reactive({ login: '', password: '' })

async function submit() {
  try {
    await auth.login(form)
    message.success('登录成功')
    router.push((route.query.redirect as string) || '/bookshelf')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '登录失败')
  }
}
</script>

<template>
  <n-form :model="form" label-placement="top" class="auth-form" @submit.prevent="submit">
    <n-form-item label="用户名或邮箱">
      <n-input v-model:value="form.login" size="large" autocomplete="username" placeholder="输入用户名或邮箱" />
    </n-form-item>
    <n-form-item label="密码">
      <n-input v-model:value="form.password" size="large" type="password" autocomplete="current-password" show-password-on="click" placeholder="输入密码" />
    </n-form-item>
    <n-button block size="large" type="primary" attr-type="submit" :loading="auth.loading">登录</n-button>
    <p class="auth-switch">还没有账号？<RouterLink to="/register">注册新用户</RouterLink></p>
  </n-form>
</template>

<style scoped>
.auth-form :deep(.n-form-item-label) {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-main);
}

.auth-form :deep(.n-input) {
  border-radius: var(--radius-medium);
  transition: all var(--transition-base);
}

.auth-form :deep(.n-input:focus-within) {
  box-shadow: 0 0 0 4px var(--color-primary-light);
}

.auth-form :deep(.n-button) {
  margin-top: var(--spacing-md);
  height: 48px;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  border-radius: var(--radius-medium);
  box-shadow: var(--shadow-button-primary);
  transition: all var(--transition-base);
}

.auth-form :deep(.n-button:hover:not(:disabled)) {
  transform: translateY(-2px);
  box-shadow: var(--shadow-button-primary-hover);
}

.auth-switch {
  margin: var(--spacing-2xl) 0 0;
  font-size: var(--font-size-base);
  color: var(--color-text-sec);
  text-align: center;
  line-height: var(--line-height-normal);
}

.auth-switch a {
  color: var(--color-primary);
  font-weight: var(--font-weight-semibold);
  text-decoration: none;
  transition: color var(--transition-base);
}

.auth-switch a:hover {
  color: var(--color-primary-hover);
  text-decoration: underline;
}
</style>
