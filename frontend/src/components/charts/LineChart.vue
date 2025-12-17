<script setup lang="ts">
import { computed } from 'vue'

type SeriesPoint = [number, number]

const props = defineProps<{
  title?: string
  subtitle?: string
  height?: string
  xAxisName?: string
  yAxisName?: string
  series: Array<{ name: string; data: SeriesPoint[] }>
  loading?: boolean
  emptyText?: string
  xMin?: number
  xMax?: number
  stacked?: boolean
}>()

const updateOptions = { notMerge: true, lazyUpdate: true } as const

const option = computed(() => {
  return {
    backgroundColor: 'transparent',
    title: {
      text: props.title ?? '',
      subtext: props.subtitle ?? '',
      left: 'left',
      textStyle: { fontWeight: 700 }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'line' }
    },
    grid: { left: 12, right: 12, top: 58, bottom: 24, containLabel: true },
    xAxis: {
      type: 'time',
      name: props.xAxisName,
      axisLabel: { hideOverlap: true },
      min: props.xMin ?? undefined,
      max: props.xMax ?? undefined
    },
    yAxis: {
      type: 'value',
      name: props.yAxisName,
      scale: true
    },
    legend: {
      top: 26,
      type: 'scroll'
    },
    series: props.series.map(s => ({
      name: s.name,
      type: 'line',
      stack: props.stacked ? 'total' : undefined,
      areaStyle: props.stacked ? { opacity: 0.25 } : undefined,
      showSymbol: false,
      smooth: false,
      connectNulls: false,
      emphasis: { focus: 'series' },
      data: s.data
    }))
  }
})
</script>

<template>
  <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-4 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
    <div v-if="loading" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />

    <div v-else-if="!series?.length || series.every(s => !s.data.length)" class="h-36 flex items-center justify-center text-sm text-slate-500 dark:text-slate-300">
      {{ emptyText ?? 'No data for selected period' }}
    </div>

    <VChart
      v-else
      class="w-full"
      :style="{ height: height ?? '320px' }"
      :option="option"
      :update-options="updateOptions"
      autoresize
    />
  </div>
</template>
