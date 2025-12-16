<script setup lang="ts">
defineProps<{
  title: string
  value: string
  hint?: string
}>()
</script>

<template>
  <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
    <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-300">{{ title }}</div>
    <div class="mt-1 text-2xl font-semibold text-slate-900 dark:text-slate-50">{{ value }}</div>
    <div v-if="hint" class="mt-1 text-xs text-slate-500 dark:text-slate-300">{{ hint }}</div>
  </div>
</template>
