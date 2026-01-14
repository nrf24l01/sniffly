import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useRangeStore } from '@/stores/range'

import { chartsService, type CountryChartResponse, type DomainChartResponse, type ProtoChartResponse, type TrafficChartResponse } from '@/service/charts'
import { tablesService, type CountryTableResponse, type DomainTableResponse, type ProtoTableResponse, type TrafficTableResponse } from '@/service/tables'

import type { RangeMode, RangePreset } from '@/types/range'

export type WidgetMode = 'chart' | 'table'

export function useDashboard() {
  const router = useRouter()
  const auth = useAuthStore()

  const rangeStore = useRangeStore()
  const { preset, rangeMode, fromExpr, toExpr, absoluteFromMs, absoluteToMs } = storeToRefs(rangeStore)

  const nowMs = () => Date.now()
  const rangeError = ref<string | null>(null)

  function presetToExpr(p: RangePreset) {
    return p === '1h' ? 'now-1h' : p === '6h' ? 'now-6h' : p === '24h' ? 'now-24h' : 'now-7d'
  }

  function parseNowExpr(exprRaw: string, baseMs: number) {
    const expr = exprRaw.trim()
    if (!expr) throw new Error('Empty time expression')

    // Allow unix seconds/milliseconds
    if (/^\d{10}$/.test(expr)) return Number(expr) * 1000
    if (/^\d{13}$/.test(expr)) return Number(expr)

    // Allow ISO-like date strings
    if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(expr)) {
      const d = new Date(expr)
      if (!Number.isFinite(d.getTime())) throw new Error(`Invalid date: ${expr}`)
      return d.getTime()
    }

    if (!expr.startsWith('now')) throw new Error(`Expression must start with "now" (got: ${expr})`)

    // Supported units:
    // y (years), m (months), w (weeks), d (days), h (hours), min (minutes), s (seconds)
    // Note: we treat "m" as months; use "min" for minutes.
    const rest = expr.slice(3)
    // IMPORTANT: put "min" before "m" so "now-2min" is not parsed as "now-2m" + "in".
    const re = /([+-])(\d+)(y|min|m|w|d|h|s)/g

    let date = new Date(baseMs)
    let match: RegExpExecArray | null
    let consumed = 0
    while ((match = re.exec(rest))) {
      consumed += match[0].length
      const sign = match[1] === '-' ? -1 : 1
      const value = Number(match[2])
      const unit = match[3]

      if (!Number.isFinite(value) || value < 0) throw new Error(`Invalid amount in ${match[0]}`)

      if (unit === 'y') date.setFullYear(date.getFullYear() + sign * value)
      else if (unit === 'm') date.setMonth(date.getMonth() + sign * value)
      else {
        const ms =
          unit === 'w'
            ? value * 7 * 24 * 3600_000
            : unit === 'd'
              ? value * 24 * 3600_000
              : unit === 'h'
                ? value * 3600_000
                : unit === 'min'
                  ? value * 60_000
                  : value * 1000
        date = new Date(date.getTime() + sign * ms)
      }
    }

    if (rest.trim().length !== consumed) {
      // Allow empty rest
      if (rest.trim().length !== 0) throw new Error(`Invalid expression tail: ${rest}`)
    }

    return date.getTime()
  }

  function setRange(fromMs: number, toMs: number) {
    if (!Number.isFinite(fromMs) || !Number.isFinite(toMs)) throw new Error('Invalid range')
    if (fromMs >= toMs) throw new Error('from must be < to')
    range.value = { fromMs, toMs }
  }

  const range = ref({ fromMs: nowMs() - 24 * 3600_000, toMs: nowMs() })

  function applyExprRange() {
    rangeError.value = null
    try {
      const base = nowMs()
      const fromMs = parseNowExpr(fromExpr.value, base)
      const toMs = parseNowExpr(toExpr.value, base)
      setRange(fromMs, toMs)
      rangeMode.value = 'expr'
      absoluteFromMs.value = fromMs
      absoluteToMs.value = toMs
    } catch (e: any) {
      rangeError.value = e?.message ?? String(e)
    }
  }

  function applyAbsoluteRange(fromMs: number, toMs: number) {
    rangeError.value = null
    try {
      setRange(fromMs, toMs)
      rangeMode.value = 'absolute'
      absoluteFromMs.value = fromMs
      absoluteToMs.value = toMs
    } catch (e: any) {
      rangeError.value = e?.message ?? String(e)
    }
  }

  // Initialize range from persisted Pinia state.
  ;(() => {
    if (rangeMode.value === 'absolute' && absoluteFromMs.value != null && absoluteToMs.value != null) {
      applyAbsoluteRange(absoluteFromMs.value, absoluteToMs.value)
      return
    }

    if (rangeMode.value === 'expr') {
      applyExprRange()
      return
    }

    // preset (default)
    fromExpr.value = presetToExpr(preset.value)
    toExpr.value = 'now'
    applyExprRange()
    rangeMode.value = 'preset'
  })()

  // keep presets as shortcuts
  watch(preset, p => {
    fromExpr.value = presetToExpr(p)
    toExpr.value = 'now'
    applyExprRange()
    rangeMode.value = 'preset'
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
  const trafficChart = ref<TrafficChartResponse | null>(null)
  const domainsChart = ref<DomainChartResponse | null>(null)
  const countriesChart = ref<CountryChartResponse | null>(null)
  const protosChart = ref<ProtoChartResponse | null>(null)

  const trafficTable = ref<TrafficTableResponse | null>(null)
  const domainsTable = ref<DomainTableResponse | null>(null)
  const countriesTable = ref<CountryTableResponse | null>(null)
  const protosTable = ref<ProtoTableResponse | null>(null)

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
    rangeMode,
    rangeError,
    fromExpr,
    toExpr,
    absoluteFromMs,
    absoluteToMs,
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

    // state
    loadingCharts,
    loadingTables,
    error,

    // actions
    loadCharts,
    ensureTable,
    applyExprRange,
    applyAbsoluteRange
  }
}
