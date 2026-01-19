<script setup lang="ts">
const props = defineProps<{
  modelValue: 'chart' | 'table'
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: 'chart' | 'table'): void
}>()

function set(v: 'chart' | 'table') {
  if (props.disabled) return
  emit('update:modelValue', v)
}
</script>

<template>
  <div class="inline-flex rounded-xl border border-slate-200/70 bg-white/70 p-1 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
    <button
      type="button"
      class="px-3 py-1.5 text-xs font-semibold rounded-lg transition"
      :class="modelValue === 'chart' ? 'bg-slate-900 text-white dark:bg-slate-50 dark:text-slate-900' : 'text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800/60'"
      :disabled="disabled"
      @click="set('chart')"
    >
График
    </button>
    <button
      type="button"
      class="px-3 py-1.5 text-xs font-semibold rounded-lg transition"
      :class="modelValue === 'table' ? 'bg-slate-900 text-white dark:bg-slate-50 dark:text-slate-900' : 'text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800/60'"
      :disabled="disabled"
      @click="set('table')"
    >
      Таблица
    </button>
  </div>
</template>
