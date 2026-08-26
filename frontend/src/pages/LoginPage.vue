<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, User, Watermelon } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/utils/errors'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({ username: 'admin', password: 'admin123' })

async function submit() {
  try {
    await auth.login(form.username, form.password)
    await router.replace('/ponds')
  } catch (error) {
    ElMessage.error(errorMessage(error))
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-brand">
      <div class="login-brand-content">
        <span class="brand-symbol"><Watermelon /></span>
        <p class="eyebrow">规模化养殖作业系统</p>
        <h1>水域智控</h1>
        <p>把水质变化、投喂决策与现场执行纳入同一条可追溯链路。</p>
        <div class="water-lines"><span /><span /><span /></div>
      </div>
    </section>
    <section class="login-panel">
      <form class="login-form" @submit.prevent="submit">
        <div><p class="eyebrow">AQUA COMMAND</p><h2>登录控制台</h2><p class="muted">使用已分配的场区账号继续</p></div>
        <label>用户名<el-input v-model="form.username" size="large" :prefix-icon="User" autocomplete="username" /></label>
        <label>密码<el-input v-model="form.password" size="large" type="password" :prefix-icon="Lock" show-password autocomplete="current-password" /></label>
        <el-button type="primary" size="large" native-type="submit" :loading="auth.loading">进入工作台</el-button>
        <p class="demo-hint">演示账号：admin / admin123</p>
      </form>
    </section>
  </main>
</template>
