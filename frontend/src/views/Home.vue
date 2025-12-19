<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import LineChart from '@/components/charts/LineChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import StatCard from '@/components/ui/StatCard.vue'
import ViewSwitch from '@/components/ui/ViewSwitch.vue'

import { useDashboard, type WidgetMode } from '@/composables/useDashboard'
import type { RangePreset } from '@/types/range'
import { useDashboardDerived } from '@/composables/useDashboardDerived'
import { formatBytes, formatNumber, formatDateTime } from '@/utils/format'

const {
  preset,
  range,
  rangeError,
  fromExpr,
  toExpr,
  absoluteFromMs,
  absoluteToMs,
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
  ensureTable,
  applyExprRange,
  applyAbsoluteRange
} = useDashboard()

const presets = ref<RangePreset[]>(['1h', '6h', '24h', '7d'])
const customOpen = ref(false)
const selectedDevice = computed(() => devices.value.find(d => d.mac === selectedMac.value) ?? null)

function toDatetimeLocal(ms: number) {
  const d = new Date(ms)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function parseDatetimeLocal(value: string) {
  // value: YYYY-MM-DDTHH:mm
  const m = value.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/)
  if (!m) return null
  const year = Number(m[1])
  const month = Number(m[2]) - 1
  const day = Number(m[3])
  const hour = Number(m[4])
  const minute = Number(m[5])
  const dt = new Date(year, month, day, hour, minute, 0, 0)
  const t = dt.getTime()
  return Number.isFinite(t) ? t : null
}

const fromLocal = ref('')
const toLocal = ref('')

watch(
  range,
  r => {
    fromLocal.value = toDatetimeLocal(r.fromMs)
    toLocal.value = toDatetimeLocal(r.toMs)
  },
  { immediate: true }
)

watch(rangeError, e => {
  if (e) customOpen.value = true
})

function applyAbsoluteFromInputs() {
  const fromMs = parseDatetimeLocal(fromLocal.value)
  const toMs = parseDatetimeLocal(toLocal.value)
  if (fromMs == null || toMs == null) return
  // Keep text inputs in sync with calendar selection (parser accepts YYYY-MM-DDTHH:mm).
  fromExpr.value = fromLocal.value
  toExpr.value = toLocal.value
  applyAbsoluteRange(fromMs, toMs)
}

const {
  topDomainsChart,
  topCountriesChart,
  topProtosChart,
  trafficSeries,
  domainsTimelineSeries,
  countriesTimelineSeries,
  protosTimelineSeries,
  companiesTimelineSeries,
  domainsRowsTable,
  countriesRowsTable,
  protosRowsTable
} = useDashboardDerived({
  selectedMac,
  trafficChart,
  domainsChart,
  countriesChart,
  protosChart,
  domainsTable,
  countriesTable,
  protosTable
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
    <div class="mx-auto w-full max-w-none px-4 py-6 sm:px-6 lg:px-8">
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
              @click="preset = p; customOpen = false"
            >
              {{ p }}
            </button>

            <button
              class="px-3 py-1.5 text-sm font-medium rounded-lg transition"
              :class="customOpen ? 'bg-slate-900 text-white shadow-sm dark:bg-slate-50 dark:text-slate-900' : 'text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800/60'"
              @click="customOpen = true"
            >
              Своё
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

      <section v-if="customOpen" class="mt-4 grid gap-3 md:grid-cols-2">
        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-300">Range (Grafana-style)</div>
          <div class="mt-2 grid gap-2 sm:grid-cols-2">
            <input
              v-model="fromExpr"
              class="w-full rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm text-slate-900 shadow-sm backdrop-blur outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50"
              placeholder="from: now-24h"
            />
            <input
              v-model="toExpr"
              class="w-full rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm text-slate-900 shadow-sm backdrop-blur outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50"
              placeholder="to: now"
            />
          </div>
          <div class="mt-2 flex items-center justify-between gap-2">
            <div class="text-xs text-slate-500 dark:text-slate-300">Units: y, m(month), w, d, h, min, s</div>
            <button
              class="inline-flex items-center justify-center rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm font-semibold text-slate-900 shadow-sm backdrop-blur transition hover:bg-white dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50 dark:hover:bg-slate-900"
              :disabled="refreshing"
              @click="applyExprRange"
            >
              Apply
            </button>
          </div>
        </div>

        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-300">Range (Calendar)</div>
          <div class="mt-2 grid gap-2 sm:grid-cols-2">
            <input
              v-model="fromLocal"
              type="datetime-local"
              class="w-full rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm text-slate-900 shadow-sm backdrop-blur outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50"
            />
            <input
              v-model="toLocal"
              type="datetime-local"
              class="w-full rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm text-slate-900 shadow-sm backdrop-blur outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50"
            />
          </div>
          <div class="mt-2 flex items-center justify-end">
            <button
              class="inline-flex items-center justify-center rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm font-semibold text-slate-900 shadow-sm backdrop-blur transition hover:bg-white dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50 dark:hover:bg-slate-900"
              :disabled="refreshing"
              @click="applyAbsoluteFromInputs"
            >
              Apply
            </button>
          </div>
        </div>
      </section>

      <section
        v-if="customOpen && rangeError"
        class="mt-3 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200"
      >
        {{ rangeError }}
      </section>

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

      <section class="mt-6 grid gap-4 md:grid-cols-2">
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

        <div class="rounded-2xl border border-slate-200/70 bg-white/70 p-3 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
          <div class="flex items-start justify-between gap-3 px-1 pb-2">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Companies</div>
              <div class="text-xs text-slate-500 dark:text-slate-300">Chart: per-company requests over time (from /charts/countries)</div>
            </div>
          </div>

          <LineChart
            title=""
            subtitle=""
            xAxisName="time"
            yAxisName="requests"
            :series="companiesTimelineSeries"
            :loading="loadingCharts"
            :stacked="true"
            :xMin="range.fromMs"
            :xMax="range.toMs"
            height="320px"
          />
        </div>
      </section>

      <section class="mt-4">
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
      </section>
    </div>
  </main>
</template>