<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title?: string
  subtitle?: string
  height?: string
  xAxisName?: string
  yAxisName?: string
  categories: string[]
  values: number[]
  loading?: boolean
  emptyText?: string
}>()

const option = computed(() => {
  return {
    backgroundColor: 'transparent',
    title: {
      text: props.title ?? '',
      subtext: props.subtitle ?? '',
      left: 'left',
      textStyle: { fontWeight: 700 }
    },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 12, right: 12, top: 58, bottom: 24, containLabel: true },
    xAxis: {
      type: 'category',
      data: props.categories,
      name: props.xAxisName,
      axisLabel: { interval: 0, rotate: props.categories.length > 8 ? 35 : 0 }
    },
    yAxis: {
      type: 'value',
      name: props.yAxisName
    },
    series: [
      {
        type: 'bar',
        data: props.values,
        barMaxWidth: 36,
        itemStyle: { borderRadius: [10, 10, 6, 6] }
      }
    ]
  }
})
</script>

<template>
  <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
    <div v-if="loading" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />

    <div v-else-if="!categories.length || !values.length" class="h-36 flex items-center justify-center text-sm text-slate-500 dark:text-slate-300">
      {{ emptyText ?? 'No data for selected period' }}
    </div>

    <VChart v-else class="w-full" :style="{ height: height ?? '320px' }" :option="option" autoresize />
  </div>
</template>
