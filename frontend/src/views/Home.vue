<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { ArrowPathIcon, PencilSquareIcon, XMarkIcon } from '@heroicons/vue/24/outline'

import LineChart from '@/components/charts/LineChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import StatCard from '@/components/ui/StatCard.vue'
import ViewSwitch from '@/components/ui/ViewSwitch.vue'

import { useDashboard, type WidgetMode } from '@/composables/useDashboard'
import type { RangePreset } from '@/types/range'
import { useDashboardDerived } from '@/composables/useDashboardDerived'
import { formatBytes, formatNumber, formatDateTime } from '@/utils/format'
import type { DeviceListItem } from '@/service/devices'

const {
  preset,
  range,
  rangeError,
  fromExpr,
  toExpr,
  absoluteFromMs,
  absoluteToMs,
  devices,
  loadingDevices,
  devicesError,
  selectedDeviceIds,
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
  loadingCharts,
  loadingTables,
  error,
  loadCharts,
  ensureTable,
  loadDevices,
  updateDeviceLabel,
  applyExprRange,
  applyAbsoluteRange
} = useDashboard()

const presets = ref<RangePreset[]>(['1h', '6h', '24h', '7d'])
const customOpen = ref(false)

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
  trafficChart,
  domainsChart,
  countriesChart,
  protosChart,
  domainsTable,
  countriesTable,
  protosTable
})

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

const allDevices = computed(() => (selectedDeviceIds.value?.length ?? 0) === 0)

const sortedDevices = computed(() => {
  const arr = Array.isArray(devices.value) ? devices.value : []
  return [...arr].sort((a, b) => {
    const an = (a.user_label ?? '').trim() || a.mac
    const bn = (b.user_label ?? '').trim() || b.mac
    return an.localeCompare(bn)
  })
})

const devicesHint = computed(() => {
  if (allDevices.value) return 'Показываются все устройства'
  return `Выбрано: ${(selectedDeviceIds.value?.length ?? 0)} из ${sortedDevices.value.length}`
})

function isSelectedDevice(id: string) {
  return (selectedDeviceIds.value ?? []).includes(id)
}

function toggleAllDevices(next: boolean) {
  if (next) selectedDeviceIds.value = []
}

function toggleDevice(id: string, next: boolean) {
  const current = new Set(selectedDeviceIds.value ?? [])
  if (next) current.add(id)
  else current.delete(id)
  selectedDeviceIds.value = Array.from(current)
}

const showEditDevice = ref(false)
const selectedDevice = ref<DeviceListItem | null>(null)
const editLabel = ref('')
const editWorking = ref(false)
const editError = ref<string | null>(null)

function openEditDevice(d: DeviceListItem) {
  selectedDevice.value = d
  editLabel.value = d.user_label ?? ''
  editError.value = null
  showEditDevice.value = true
}

function closeEditDevice() {
  showEditDevice.value = false
  selectedDevice.value = null
  editLabel.value = ''
  editError.value = null
}

async function saveEditDevice() {
  if (!selectedDevice.value) return
  editWorking.value = true
  editError.value = null
  try {
    await updateDeviceLabel(selectedDevice.value.uuid, editLabel.value)
    closeEditDevice()
  } catch (e: any) {
    editError.value = e?.response?.data?.message ?? e?.message ?? String(e)
  } finally {
    editWorking.value = false
  }
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

      <section class="mt-4 rounded-2xl border border-slate-200/70 bg-white/70 p-4 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/50">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Устройства</div>
            <div class="mt-0.5 text-xs text-slate-500 dark:text-slate-300">{{ devicesHint }}</div>
          </div>

          <div class="flex items-center gap-2">
            <button
              class="inline-flex items-center gap-2 rounded-xl border border-slate-200/70 bg-white/70 px-3 py-2 text-sm font-semibold text-slate-900 shadow-sm backdrop-blur transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50 dark:hover:bg-slate-900"
              :disabled="loadingDevices"
              @click="loadDevices"
            >
              <ArrowPathIcon class="h-4 w-4" :class="loadingDevices ? 'animate-spin' : ''" />
              {{ loadingDevices ? 'Загрузка…' : 'Обновить' }}
            </button>
          </div>
        </div>

        <div v-if="devicesError" class="mt-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
          {{ devicesError }}
        </div>

        <details class="mt-3">
          <summary class="cursor-pointer select-none rounded-xl border border-slate-200/70 bg-white/60 px-3 py-2 text-sm font-semibold text-slate-900 hover:bg-white dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-50 dark:hover:bg-slate-900">
            Выбрать устройства
          </summary>

          <div class="mt-3 flex flex-wrap items-center gap-3">
            <label class="flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-200">
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500"
                :checked="allDevices"
                @change="toggleAllDevices(($event.target as HTMLInputElement).checked)"
              />
              Все устройства
            </label>
            <span class="text-xs text-slate-500 dark:text-slate-300">(если ничего не выбрано — тоже все)</span>
          </div>

          <div v-if="sortedDevices.length" class="mt-3 max-h-72 overflow-auto rounded-xl border border-slate-200/70 bg-white/60 dark:border-slate-800 dark:bg-slate-900/40">
            <div
              v-for="d in sortedDevices"
              :key="d.uuid"
              class="flex items-start justify-between gap-3 border-b border-slate-200/70 px-3 py-2 last:border-b-0 dark:border-slate-800"
            >
              <label class="flex flex-1 cursor-pointer items-start gap-3">
                <input
                  type="checkbox"
                  class="mt-1 h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500"
                  :checked="isSelectedDevice(d.uuid)"
                  @change="toggleDevice(d.uuid, ($event.target as HTMLInputElement).checked)"
                />
                <div class="min-w-0">
                  <div class="truncate text-sm font-semibold text-slate-900 dark:text-slate-50">{{ d.user_label || d.mac }}</div>
                  <div class="mt-0.5 truncate text-xs text-slate-500 dark:text-slate-300">{{ d.mac }} • {{ d.ip }}</div>
                </div>
              </label>

              <button
                class="inline-flex shrink-0 items-center gap-1 rounded-lg border border-slate-200/70 px-2 py-1 text-xs font-semibold text-slate-800 transition hover:bg-slate-100 dark:border-slate-800 dark:text-slate-100 dark:hover:bg-slate-800"
                @click="openEditDevice(d)"
                title="Изменить метку"
              >
                <PencilSquareIcon class="h-4 w-4" />
                Метка
              </button>
            </div>
          </div>

          <div v-else class="mt-3 text-sm text-slate-500 dark:text-slate-300">
            {{ loadingDevices ? 'Загрузка списка устройств…' : 'Устройства не найдены.' }}
          </div>
        </details>
      </section>

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

      <!-- <section class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Devices" :value="formatNumber(deviceCount)" hint="Devices with traffic in range" />
        <StatCard title="Total Up" :value="formatBytes(totalUp)" hint="Shown when Traffic table is loaded" />
        <StatCard title="Total Down" :value="formatBytes(totalDown)" hint="Shown when Traffic table is loaded" />
        <StatCard
          title="Selected"
          :value="selectedDevice ? (selectedDevice.label || selectedDevice.mac) : '—'"
          :hint="selectedDevice ? `${selectedDevice.ip} • ${selectedDevice.hostname}` : 'Pick a device to inspect'"
        />
      </section> -->

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

          <div v-else class="mt-2 h-80 overflow-auto">
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

          <div v-else class="mt-2 h-80 overflow-auto">
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

          <div v-else class="mt-2 h-80 overflow-auto">
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

          <div v-else class="mt-2 h-80 overflow-auto">
            <div v-if="loadingTables.traffic" class="h-36 w-full animate-pulse rounded-xl bg-slate-200/60 dark:bg-slate-700/40" />
            <table v-else class="min-w-full text-left text-sm">
              <thead class="text-xs uppercase text-slate-500 dark:text-slate-300">
                <tr>
                  <th class="py-2 pr-4">Up</th>
                  <th class="py-2 pr-4">Down</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
                <tr v-if="trafficTable" class="hover:bg-slate-50/80 dark:hover:bg-slate-800/30">
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatBytes(trafficTable.stats.up_bytes) }}</td>
                  <td class="py-3 pr-4 text-slate-700 dark:text-slate-200">{{ formatBytes(trafficTable.stats.down_bytes) }}</td>
                </tr>
                <tr v-if="!loadingTables.traffic && !trafficTable">
                  <td colspan="2" class="py-6 text-center text-sm text-slate-500 dark:text-slate-300">No data</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <div v-if="showEditDevice" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 px-4">
        <div class="w-full max-w-lg rounded-2xl border border-slate-200/70 bg-white p-6 shadow-2xl dark:border-slate-800 dark:bg-slate-900">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Изменить метку устройства</h2>
              <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">
                {{ selectedDevice ? `${selectedDevice.mac} • ${selectedDevice.ip}` : '' }}
              </p>
            </div>

            <button
              class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-slate-50"
              @click="closeEditDevice"
              aria-label="Close"
            >
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>

          <div v-if="editError" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
            {{ editError }}
          </div>

          <div class="mt-4">
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">Метка (user_label)</label>
            <input
              v-model="editLabel"
              type="text"
              class="mt-1 w-full rounded-xl border border-slate-200/70 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50"
              placeholder="Например, Office PC"
            />
          </div>

          <div class="mt-6 flex justify-end gap-3">
            <button
              class="rounded-xl px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800"
              :disabled="editWorking"
              @click="closeEditDevice"
            >
              Отмена
            </button>
            <button
              class="inline-flex items-center justify-center gap-2 rounded-xl bg-green-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-green-700 disabled:cursor-not-allowed disabled:bg-green-400"
              :disabled="editWorking || !editLabel.trim()"
              @click="saveEditDevice"
            >
              <ArrowPathIcon v-if="editWorking" class="h-4 w-4 animate-spin" />
              <span>Сохранить</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>