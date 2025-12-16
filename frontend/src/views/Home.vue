<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import LineChart from '@/components/charts/LineChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import StatCard from '@/components/ui/StatCard.vue'
import ViewSwitch from '@/components/ui/ViewSwitch.vue'

import { useDashboard, type RangePreset, type WidgetMode } from '@/composables/useDashboard'
import { formatBytes, formatNumber, formatDateTime } from '@/utils/format'

const {
  preset,
  range,
  trafficMode,
  domainsMode,
  countriesMode,
  protosMode,
  trafficChart,
  domainsChart,
  countriesChart,
  protosChart,
  trafficTable,
  domainsTable,
  countriesTable,
  protosTable,
  selectedMac,
  devices,
  loadingCharts,
  loadingTables,
  error,
  loadCharts,
  ensureTable
} = useDashboard()

const presets = ref<RangePreset[]>(['1h', '6h', '24h', '7d'])
const selectedDevice = computed(() => devices.value.find(d => d.mac === selectedMac.value) ?? null)

// Chart data builders (/charts)
// trafficSeries declared later after padding helper

const topDomainsChart = computed(() => {
  const mac = selectedMac.value
  const buckets = domainsChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  const agg = new Map<string, number>()
  for (const b of buckets) {
    for (const [domain, count] of Object.entries(b.domains ?? {})) {
      agg.set(domain, (agg.get(domain) ?? 0) + (count ?? 0))
    }
  }
  return Array.from(agg.entries())
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 12)
})

const topCountriesChart = computed(() => {
  const mac = selectedMac.value
  const buckets = countriesChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  const agg = new Map<string, number>()
  for (const b of buckets) {
    const countries = b.countries
    if (Array.isArray(countries)) {
      for (const c of countries) agg.set(c, (agg.get(c) ?? 0) + 1)
    } else if (countries && typeof countries === 'object') {
      for (const [c, count] of Object.entries(countries as Record<string, number | undefined>)) {
        agg.set(c, (agg.get(c) ?? 0) + (count ?? 0))
      }
    }
  }
  return Array.from(agg.entries())
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 10)
})

const topProtosChart = computed(() => {
  const mac = selectedMac.value
  const buckets = protosChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  const agg = new Map<string, number>()
  for (const b of buckets) {
    const protos = b.protos
    if (Array.isArray(protos)) {
      for (const p of protos) agg.set(p, (agg.get(p) ?? 0) + 1)
    } else if (protos && typeof protos === 'object') {
      for (const [p, count] of Object.entries(protos as Record<string, number | undefined>)) {
        agg.set(p, (agg.get(p) ?? 0) + (count ?? 0))
      }
    }
  }
  return Array.from(agg.entries())
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 10)
})

function buildSeriesFromBuckets(
  buckets: Array<{ bucket: number }>,
  entries: (bucket: any) => Array<[string, number]>,
  topN = 6
): Array<{ name: string; data: [number, number][] }> {
  const totals = new Map<string, number>()
  const bucketMaps = new Map<number, Map<string, number>>()

  for (const b of buckets) {
    const map = new Map<string, number>()
    for (const [k, v] of entries(b)) {
      const num = Number(v ?? 0)
      map.set(k, num)
      totals.set(k, (totals.get(k) ?? 0) + num)
    }
    bucketMaps.set(b.bucket, map)
  }

  const topKeys = Array.from(totals.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, topN)
    .map(([k]) => k)

  const allBuckets = new Set<number>()
  for (const b of buckets) allBuckets.add(b.bucket)
  allBuckets.add(Math.floor(range.value.fromMs / 1000))
  allBuckets.add(Math.floor(range.value.toMs / 1000))
  const timeline = Array.from(allBuckets.values()).sort((a, b) => a - b)

  return topKeys.map(key => ({
    name: key,
    data: timeline.map(t => {
      const val = bucketMaps.get(t)?.get(key) ?? 0
      return [t * 1000, val] as [number, number]
    })
  }))
}

const domainsTimelineSeries = computed<Array<{ name: string; data: [number, number][] }>>(() => {
  const mac = selectedMac.value
  const buckets = domainsChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  return buildSeriesFromBuckets(buckets, b => Object.entries(b.domains ?? {}).map(([k, v]) => [k, Number(v ?? 0)]))
})

const countriesTimelineSeries = computed<Array<{ name: string; data: [number, number][] }>>(() => {
  const mac = selectedMac.value
  const buckets = countriesChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  return buildSeriesFromBuckets(
    buckets,
    b => {
      const arr = Array.isArray(b.countries) ? b.countries : Object.keys(b.countries ?? {})
      return arr.map((k: string) => [k, 1])
    },
    8
  )
})

const protosTimelineSeries = computed<Array<{ name: string; data: [number, number][] }>>(() => {
  const mac = selectedMac.value
  const buckets = protosChart.value?.find(x => x.device.mac === mac)?.stats ?? []
  return buildSeriesFromBuckets(
    buckets,
    b => {
      const arr = Array.isArray(b.protos) ? b.protos : Object.keys(b.protos ?? {})
      return arr.map((k: string) => [k, 1])
    },
    8
  )
})

// Ensure traffic series also spans full range
function padPoints(points: Array<[number, number]>): Array<[number, number]> {
  const fromMs = range.value.fromMs
  const toMs = range.value.toMs
  const pts = [...points].sort((a, b) => a[0] - b[0]) as Array<[number, number]>
  if (pts.length === 0) return [[fromMs, 0], [toMs, 0]]
  const first = pts[0]
  const last = pts[pts.length - 1]
  if (first && first[0] > fromMs) pts.unshift([fromMs, 0])
  if (last && last[0] < toMs) pts.push([toMs, 0])
  return pts
}

// update trafficSeries to pad
const trafficSeries = computed<Array<{ name: string; data: [number, number][] }>>(() => {
  const mac = selectedMac.value
  if (mac === null || mac === undefined) return []
  const item = trafficChart.value.find(x => x.device.mac === mac)
  if (!item) return []

  const up: Array<[number, number]> = []
  const down: Array<[number, number]> = []
  for (const b of item.stats) {
    const t = b.bucket * 1000
    up.push([t, b.up_bytes])
    down.push([t, b.down_bytes])
  }

  return [
    { name: 'Up', data: padPoints(up) },
    { name: 'Down', data: padPoints(down) }
  ]
})

// Table helpers (/tables)
function topFromTable(stats: Record<string, number> | undefined, limit = 30) {
  if (!stats) return [] as Array<{ key: string; value: number }>
  return Object.entries(stats)
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, limit)
}

const domainsRowsTable = computed(() => {
  const mac = selectedMac.value
  const stats = domainsTable.value?.find(x => x.device.mac === mac)?.stats
  return topFromTable(stats, 30)
})
const countriesRowsTable = computed(() => {
  const mac = selectedMac.value
  const stats = countriesTable.value?.find(x => x.device.mac === mac)?.stats
  return topFromTable(stats, 30)
})
const protosRowsTable = computed(() => {
  const mac = selectedMac.value
  const stats = protosTable.value?.find(x => x.device.mac === mac)?.stats
  return topFromTable(stats, 30)
})

// Stats cards
const deviceCount = computed(() => devices.value.length)
const totalUp = computed(() => (trafficTable.value ?? []).reduce((acc, row) => acc + (row.stats?.up_bytes ?? 0), 0))
const totalDown = computed(() => (trafficTable.value ?? []).reduce((acc, row) => acc + (row.stats?.down_bytes ?? 0), 0))

const rangeLabel = computed(() => `${formatDateTime(range.value.fromMs)} → ${formatDateTime(range.value.toMs)}`)
const refreshing = computed(() => loadingCharts.value)

// Lazy table fetch when switching to table mode
watch(trafficMode, m => m === 'table' && ensureTable('traffic'))
watch(domainsMode, m => m === 'table' && ensureTable('domains'))
watch(countriesMode, m => m === 'table' && ensureTable('countries'))
watch(protosMode, m => m === 'table' && ensureTable('protos'))

function onSwitch(widget: 'traffic' | 'domains' | 'countries' | 'protos', m: WidgetMode) {
  if (m === 'table') return ensureTable(widget)
}
</script>

<template>
  <main class="h-full overflow-auto bg-gradient-to-b from-green-50 to-white dark:from-slate-950 dark:to-slate-900">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <header class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight text-slate-900 dark:text-slate-50">Dashboard</h1>
          <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">{{ rangeLabel }}</p>
        </div>

        <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div class="inline-flex rounded-xl border border-slate-200/70 bg-white/70 p-1 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
            <button
              v-for="p in presets"
              :key="p"
              class="px-3 py-1.5 text-sm font-medium rounded-lg transition"
              :class="preset === p ? 'bg-green-600 text-white shadow-sm' : 'text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800/60'"
              @click="preset = p"
            >
              {{ p }}
            </button>
          </div>

          <button
            class="inline-flex items-center justify-center rounded-xl border border-slate-200/70 bg-white/70 px-4 py-2 text-sm font-semibold text-slate-900 shadow-sm backdrop-blur transition hover:bg-white dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50 dark:hover:bg-slate-900"
            :disabled="refreshing"
            @click="loadCharts"
          >
            {{ refreshing ? 'Refreshing…' : 'Refresh' }}
          </button>
        </div>
      </header>

      <section
        v-if="error"
        class="mt-5 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200"
      >
        {{ error }}
      </section>

      <section class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Devices" :value="formatNumber(deviceCount)" hint="Devices with traffic in range" />
        <StatCard title="Total Up" :value="formatBytes(totalUp)" hint="Shown when Traffic table is loaded" />
        <StatCard title="Total Down" :value="formatBytes(totalDown)" hint="Shown when Traffic table is loaded" />
        <StatCard
          title="Selected"
          :value="selectedDevice ? (selectedDevice.label || selectedDevice.mac) : '—'"
          :hint="selectedDevice ? `${selectedDevice.ip} • ${selectedDevice.hostname}` : 'Pick a device to inspect'"
        />
      </section>

      <section class="mt-6 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="text-sm font-medium text-slate-700 dark:text-slate-200">Device</div>
        <div class="w-full lg:max-w-xl">
          <select
            v-model="selectedMac"
            class="w-full rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm text-slate-900 shadow-sm backdrop-blur outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50"
          >
            <option v-for="d in devices" :key="d.mac" :value="d.mac">{{ d.label || d.mac }} — {{ d.ip }}</option>
          </select>
        </div>
      </section>

      <section class="mt-6 grid gap-4 lg:grid-cols-2">
        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="flex items-start justify-between gap-3 px-1 pb-2">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Traffic</div>
              <div class="text-xs text-slate-500 dark:text-slate-300">Chart: time vs bytes from /charts; Table: totals from /tables</div>
            </div>
            <ViewSwitch v-model="trafficMode" :disabled="refreshing" @update:modelValue="onSwitch('traffic', $event)" />
          </div>

          <LineChart
            v-if="trafficMode === 'chart'"
            title=""
            subtitle=""
            xAxisName="time"
            yAxisName="bytes"
            :series="trafficSeries"
            :loading="loadingCharts"
            :stacked="true"
              :xMin="range.fromMs"
              :xMax="range.toMs"
            height="320px"
          />

          <div v-else class="mt-2 overflow-auto">
            <div v-if="loadingTables.traffic" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />
            <table v-else class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-300">
                <tr>
                  <th class="py-2 pr-4">Device</th>
                  <th class="py-2 pr-4">IP</th>
                  <th class="py-2 pr-4">Up</th>
                  <th class="py-2 pr-4">Down</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
                <tr v-for="row in (trafficTable ?? [])" :key="row.device.mac" class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30">
                  <td class="py-3 pr-4 font-medium text-slate-900 dark:text-slate-50">{{ row.device.label || row.device.mac }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ row.device.ip }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatBytes(row.stats.up_bytes) }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatBytes(row.stats.down_bytes) }}</td>
                </tr>
                <tr v-if="!loadingTables.traffic && (!trafficTable || trafficTable.length === 0)">
                  <td colspan="4" class="py-6 text-center text-sm text-slate-500 dark:text-slate-300">No data</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="flex items-start justify-between gap-3 px-1 pb-2">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Domains</div>
              <div class="text-xs text-slate-500 dark:text-slate-300">Chart: per-domain requests over time; Table: totals from /tables</div>
            </div>
            <ViewSwitch v-model="domainsMode" :disabled="refreshing" @update:modelValue="onSwitch('domains', $event)" />
          </div>

          <LineChart
            v-if="domainsMode === 'chart'"
            title=""
            subtitle=""
            xAxisName="time"
            yAxisName="requests"
            :series="domainsTimelineSeries"
            :loading="loadingCharts"
            :stacked="true"
              :xMin="range.fromMs"
              :xMax="range.toMs"
            height="320px"
          />

          <div v-else class="mt-2 overflow-auto">
            <div v-if="loadingTables.domains" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />
            <table v-else class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-300">
                <tr>
                  <th class="py-2 pr-4">Domain</th>
                  <th class="py-2 pr-4">Requests</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
                <tr v-for="row in domainsRowsTable" :key="row.key" class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30">
                  <td class="py-3 pr-4 font-medium text-slate-900 dark:text-slate-50">{{ row.key }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatNumber(row.value) }}</td>
                </tr>
                <tr v-if="!loadingTables.domains && (!domainsTable || domainsRowsTable.length === 0)">
                  <td colspan="2" class="py-6 text-center text-sm text-slate-500 dark:text-slate-300">No data</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="mt-4 grid gap-4 lg:grid-cols-2">
        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="flex items-start justify-between gap-3 px-1 pb-2">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Countries</div>
              <div class="text-xs text-slate-500 dark:text-slate-300">Chart: per-country requests over time; Table: totals from /tables</div>
            </div>
            <ViewSwitch v-model="countriesMode" :disabled="refreshing" @update:modelValue="onSwitch('countries', $event)" />
          </div>

          <LineChart
            v-if="countriesMode === 'chart'"
            title=""
            subtitle=""
            xAxisName="time"
            yAxisName="requests"
            :series="countriesTimelineSeries"
            :loading="loadingCharts"
            :stacked="true"
              :xMin="range.fromMs"
              :xMax="range.toMs"
            height="320px"
          />

          <div v-else class="mt-2 overflow-auto">
            <div v-if="loadingTables.countries" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />
            <table v-else class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-300">
                <tr>
                  <th class="py-2 pr-4">Country</th>
                  <th class="py-2 pr-4">Requests</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
                <tr v-for="row in countriesRowsTable" :key="row.key" class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30">
                  <td class="py-3 pr-4 font-medium text-slate-900 dark:text-slate-50">{{ row.key }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatNumber(row.value) }}</td>
                </tr>
                <tr v-if="!loadingTables.countries && (!countriesTable || countriesRowsTable.length === 0)">
                  <td colspan="2" class="py-6 text-center text-sm text-slate-500 dark:text-slate-300">No data</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="flex items-start justify-between gap-3 px-1 pb-2">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Protocols</div>
              <div class="text-xs text-slate-500 dark:text-slate-300">Chart: per-protocol requests over time; Table: totals from /tables</div>
            </div>
            <ViewSwitch v-model="protosMode" :disabled="refreshing" @update:modelValue="onSwitch('protos', $event)" />
          </div>

          <LineChart
            v-if="protosMode === 'chart'"
            title=""
            subtitle=""
            xAxisName="time"
            yAxisName="requests"
            :series="protosTimelineSeries"
            :loading="loadingCharts"
            :stacked="true"
              :xMin="range.fromMs"
              :xMax="range.toMs"
            height="320px"
          />

          <div v-else class="mt-2 overflow-auto">
            <div v-if="loadingTables.protos" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />
            <table v-else class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-300">
                <tr>
                  <th class="py-2 pr-4">Proto</th>
                  <th class="py-2 pr-4">Requests</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
                <tr v-for="row in protosRowsTable" :key="row.key" class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30">
                  <td class="py-3 pr-4 font-medium text-slate-900 dark:text-slate-50">{{ row.key }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatNumber(row.value) }}</td>
                </tr>
                <tr v-if="!loadingTables.protos && (!protosTable || protosRowsTable.length === 0)">
                  <td colspan="2" class="py-6 text-center text-sm text-slate-500 dark:text-slate-300">No data</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </div>
  </main>
</template>