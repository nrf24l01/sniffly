import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

import { chartsService, type DeviceDomainItem, type DeviceTrafficItem, type DeviceCountryItem, type DeviceProtoItem } from '@/service/charts'
import { tablesService, type DeviceTrafficSummary, type DeviceDomainSummary, type DeviceCountrySummary, type DeviceProtoSummary } from '@/service/tables'

export type RangePreset = '1h' | '6h' | '24h' | '7d'
export type WidgetMode = 'chart' | 'table'

export function useDashboard() {
  const router = useRouter()
  const auth = useAuthStore()

  const preset = ref<RangePreset>('24h')
  const nowMs = () => Date.now()

  function presetToRange(p: RangePreset) {
    const toMs = nowMs()
    const deltaMs =
      p === '1h'
        ? 3600_000
        : p === '6h'
          ? 6 * 3600_000
          : p === '24h'
            ? 24 * 3600_000
            : 7 * 24 * 3600_000
    return { fromMs: toMs - deltaMs, toMs }
  }

  const range = ref(presetToRange(preset.value))
  watch(preset, p => {
    range.value = presetToRange(p)
  })

  const fromSeconds = computed(() => Math.floor(range.value.fromMs / 1000))
  const toSeconds = computed(() => Math.floor(range.value.toMs / 1000))

  // Per-widget mode
  const trafficMode = ref<WidgetMode>('chart')
  const domainsMode = ref<WidgetMode>('chart')
  const countriesMode = ref<WidgetMode>('chart')
  const protosMode = ref<WidgetMode>('chart')

  // Loading / errors
  const loadingCharts = ref(false)
  const loadingTables = ref<{ traffic?: boolean; domains?: boolean; countries?: boolean; protos?: boolean }>({})
  const error = ref<string | null>(null)

  // Data
  const trafficChart = ref<DeviceTrafficItem[]>([])
  const domainsChart = ref<DeviceDomainItem[] | null>(null)
  const countriesChart = ref<DeviceCountryItem[] | null>(null)
  const protosChart = ref<DeviceProtoItem[] | null>(null)

  const trafficTable = ref<DeviceTrafficSummary[] | null>(null)
  const domainsTable = ref<DeviceDomainSummary[] | null>(null)
  const countriesTable = ref<DeviceCountrySummary[] | null>(null)
  const protosTable = ref<DeviceProtoSummary[] | null>(null)

  // Selection
  const selectedMac = ref<string | null>(null)
  const devices = computed(() => {
    const map = new Map<string, { mac: string; label: string; ip: string; hostname: string }>()
    for (const item of trafficChart.value) map.set(item.device.mac, item.device)
    return Array.from(map.values()).sort((a, b) => (a.label || a.mac).localeCompare(b.label || b.mac))
  })

  watch(
    devices,
    d => {
      if (selectedMac.value == null && d.length) selectedMac.value = d[0]?.mac ?? null
      if (selectedMac.value != null && !d.some(x => x.mac === selectedMac.value)) selectedMac.value = d[0]?.mac ?? null
    },
    { immediate: true }
  )

  async function loadCharts() {
    if (!auth.isAuthenticated) {
      router.replace({ name: 'Login' })
      return
    }

    loadingCharts.value = true
    error.value = null

    // Reset table caches when range changes
    trafficTable.value = null
    domainsTable.value = null
    countriesTable.value = null
    protosTable.value = null

    try {
      const [traffic, domains, countries, protos] = await Promise.all([
        chartsService.traffic(fromSeconds.value, toSeconds.value),
        chartsService.domains(fromSeconds.value, toSeconds.value),
        chartsService.countries(fromSeconds.value, toSeconds.value),
        chartsService.protos(fromSeconds.value, toSeconds.value)
      ])

      trafficChart.value = traffic
      domainsChart.value = domains
      countriesChart.value = countries
      protosChart.value = protos
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      loadingCharts.value = false
    }
  }

  async function ensureTable(kind: 'traffic' | 'domains' | 'countries' | 'protos') {
    if (!auth.isAuthenticated) {
      router.replace({ name: 'Login' })
      return
    }

    // cached
    if (kind === 'traffic' && trafficTable.value) return
    if (kind === 'domains' && domainsTable.value) return
    if (kind === 'countries' && countriesTable.value) return
    if (kind === 'protos' && protosTable.value) return

    error.value = null
    loadingTables.value = { ...loadingTables.value, [kind]: true }

    try {
      if (kind === 'traffic') trafficTable.value = await tablesService.traffic(fromSeconds.value, toSeconds.value)
      if (kind === 'domains') domainsTable.value = await tablesService.domains(fromSeconds.value, toSeconds.value)
      if (kind === 'countries') countriesTable.value = await tablesService.countries(fromSeconds.value, toSeconds.value)
      if (kind === 'protos') protosTable.value = await tablesService.protos(fromSeconds.value, toSeconds.value)
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
      throw e
    } finally {
      loadingTables.value = { ...loadingTables.value, [kind]: false }
    }
  }

  watch([fromSeconds, toSeconds], () => {
    void loadCharts()
  })

  onMounted(() => {
    void loadCharts()
  })

  return {
    // range
    preset,
    range,
    fromSeconds,
    toSeconds,

    // modes
    trafficMode,
    domainsMode,
    countriesMode,
    protosMode,

    // data
    trafficChart,
    domainsChart,
    countriesChart,
    protosChart,

    trafficTable,
    domainsTable,
    countriesTable,
    protosTable,

    // selection
    selectedMac,
    devices,

    // state
    loadingCharts,
    loadingTables,
    error,

    // actions
    loadCharts,
    ensureTable
  }
}
