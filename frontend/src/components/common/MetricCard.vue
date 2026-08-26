<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{ label: string; value: string | number; hint?: string; tone?: 'green' | 'blue' | 'amber' | 'red'; icon: Component }>()
</script>

<template>
  <div class="metric-card" :data-tone="tone || 'green'">
    <div class="metric-icon"><component :is="icon" /></div>
    <div>
      <span class="metric-label">{{ label }}</span>
      <strong>{{ value }}</strong>
      <small v-if="hint">{{ hint }}</small>
    </div>
  </div>
</template>
