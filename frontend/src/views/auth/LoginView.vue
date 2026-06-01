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
  <n-form :model="form" label-placement="top" @submit.prevent="submit">
    <n-form-item label="用户名或邮箱">
      <n-input v-model:value="form.login" autocomplete="username" placeholder="输入用户名或邮箱" />
    </n-form-item>
    <n-form-item label="密码">
      <n-input v-model:value="form.password" type="password" autocomplete="current-password" show-password-on="click" placeholder="输入密码" />
    </n-form-item>
    <n-button block type="primary" attr-type="submit" :loading="auth.loading">登录</n-button>
    <p class="auth-switch">还没有账号？<RouterLink to="/register">注册新用户</RouterLink></p>
  </n-form>
</template>

<style scoped>
.auth-switch {
  margin: 16px 0 0;
  color: var(--color-text-sec);
  text-align: center;
}

.auth-switch a {
  color: var(--color-primary);
}
</style>
