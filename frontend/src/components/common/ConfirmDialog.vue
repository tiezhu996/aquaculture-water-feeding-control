<script setup lang="ts">
const props = withDefaults(defineProps<{ modelValue: boolean; title: string; message: string; confirmText?: string; danger?: boolean; loading?: boolean }>(), {
  confirmText: '确认', danger: false, loading: false,
})
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; confirm: [] }>()
</script>

<template>
  <el-dialog :model-value="props.modelValue" :title="title" width="420px" @close="emit('update:modelValue', false)">
    <div class="confirm-message">{{ message }}</div>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button :type="danger ? 'danger' : 'primary'" :loading="loading" @click="emit('confirm')">{{ confirmText }}</el-button>
    </template>
  </el-dialog>
</template>
