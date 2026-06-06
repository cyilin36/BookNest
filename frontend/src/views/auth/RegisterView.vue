<script setup lang="ts">
import { reactive } from 'vue'
import { useMessage } from 'naive-ui'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { AppAPIError } from '@/api/client'

const auth = useAuthStore()
const router = useRouter()
const message = useMessage()
const form = reactive({ username: '', email: '', nickname: '', password: '' })

function showRegisterError(error: unknown) {
  if (error instanceof AppAPIError) {
    if (error.code === 'username_exists') {
      message.warning('用户名已存在')
      return
    }
    if (error.code === 'email_exists') {
      message.warning('邮箱已存在')
      return
    }
    if (error.code === 'validation_failed') {
      message.warning(error.message || '注册信息不符合要求')
      return
    }
  }
  message.error(error instanceof Error ? error.message : '注册失败')
}

async function submit() {
  if (form.password.length < 6) {
    message.warning('密码至少 6 位')
    return
  }

  try {
    await auth.register({
      username: form.username,
      email: form.email || null,
      nickname: form.nickname || null,
      password: form.password
    })
    message.success('注册成功')
    router.push('/bookshelf')
  } catch (error) {
    showRegisterError(error)
  }
}
</script>

<template>
  <n-form :model="form" label-placement="top" @submit.prevent="submit">
    <n-form-item label="用户名">
      <n-input v-model:value="form.username" autocomplete="username" placeholder="至少 1 个字符" />
    </n-form-item>
    <n-form-item label="邮箱">
      <n-input v-model:value="form.email" autocomplete="email" placeholder="可选" />
    </n-form-item>
    <n-form-item label="昵称">
      <n-input v-model:value="form.nickname" placeholder="可选" />
    </n-form-item>
    <n-form-item label="密码">
      <n-input v-model:value="form.password" type="password" autocomplete="new-password" show-password-on="click" placeholder="至少 6 位" />
    </n-form-item>
    <n-button block type="primary" attr-type="submit" :loading="auth.loading">注册</n-button>
    <p class="auth-switch">已有账号？<RouterLink to="/login">去登录</RouterLink></p>
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
