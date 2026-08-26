<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Collection, DataAnalysis, DocumentChecked, Fold, List, SwitchButton, Watermelon } from '@element-plus/icons-vue'
import { useAuth } from '@/hooks/useAuth'

const route = useRoute()
const router = useRouter()
const collapsed = ref(false)
const { user, canReview, logout } = useAuth()
const titles: Record<string, string> = {
  '/ponds': '养殖池工作台', '/readings': '水质读数', '/plans': '投喂计划', '/executions': '执行反馈', '/audit': '操作审计',
}
const title = computed(() => titles[route.path] || '水域智控')

function signOut() {
  logout()
  void router.replace('/login')
}
</script>

<template>
  <div class="app-shell" :class="{ collapsed }">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark"><Watermelon /></span>
        <div class="brand-copy"><strong>水域智控</strong><small>Aqua Command</small></div>
      </div>
      <nav class="main-nav" aria-label="主导航">
        <router-link to="/ponds"><Collection /><span>养殖池</span></router-link>
        <router-link to="/readings"><DataAnalysis /><span>水质读数</span></router-link>
        <router-link to="/plans"><DocumentChecked /><span>投喂计划</span></router-link>
        <router-link to="/executions"><List /><span>执行反馈</span></router-link>
        <router-link v-if="canReview()" to="/audit"><Collection /><span>操作审计</span></router-link>
      </nav>
      <div class="sidebar-footer">
        <div class="user-avatar">{{ user?.displayName?.slice(0, 1) }}</div>
        <div class="user-copy"><strong>{{ user?.displayName }}</strong><small>{{ user?.role }}</small></div>
        <el-tooltip content="退出登录" placement="top"><button class="icon-button" @click="signOut"><SwitchButton /></button></el-tooltip>
      </div>
    </aside>
    <main class="main-area">
      <header class="topbar">
        <button class="icon-button collapse-button" title="收起导航" @click="collapsed = !collapsed"><Fold /></button>
        <div><span class="breadcrumb">规模化养殖控制</span><h1>{{ title }}</h1></div>
        <div class="system-state"><span />系统在线</div>
      </header>
      <div class="page-container"><router-view /></div>
    </main>
  </div>
</template>
