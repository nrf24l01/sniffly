import { computed, type Ref } from 'vue'

import type {
  DeviceCountryItem,
  DeviceDomainItem,
  DeviceProtoItem,
  DeviceTrafficItem
} from '@/service/charts'
import type { DeviceCountrySummary, DeviceDomainSummary, DeviceProtoSummary } from '@/service/tables'

type TopRow = { key: string; value: number }
export type TimeSeries = { name: string; data: [number, number][] }

// ECharts keeps internal copies of data; large series-count × point-count explodes memory.
// Keep these fairly conservative to avoid multi-GB tabs.
const MAX_TIMELINE_POINTS = 1200
const MAX_TOTAL_POINTS = 40_000

function topFromTable(stats: Record<string, number> | undefined, limit = 30): TopRow[] {
  if (!stats) return []
  return Object.entries(stats)
    .map(([key, value]) => ({ key, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, limit)
}

function fillInternalGapsWithZero(points: Array<[number, number]>): Array<[number, number]> {
  const pts = [...points].sort((a, b) => a[0] - b[0]) as Array<[number, number]>
  if (pts.length <= 1) return pts

  let step: number | null = null
  const diffs: number[] = []
  for (let i = 1; i < pts.length; i++) {
    const d = pts[i]![0] - pts[i - 1]![0]
    if (d > 0) diffs.push(d)
  }

  if (!diffs.length) return pts

  // Choose the most common diff to avoid tiny outliers exploding the timeline.
  const freq = new Map<number, number>()
  for (const d of diffs) freq.set(d, (freq.get(d) ?? 0) + 1)
  step = Array.from(freq.entries()).sort((a, b) => b[1] - a[1])[0]?.[0] ?? null
  if (!step) return pts

  const baseStep = step

  const values = new Map<number, number>(pts.map(([t, v]) => [t, v]))
  const start = pts[0]![0]
  const end = pts[pts.length - 1]![0]

  const predicted = Math.floor((end - start) / baseStep) + 1
  const factor = predicted > MAX_TIMELINE_POINTS ? Math.ceil(predicted / MAX_TIMELINE_POINTS) : 1
  step = baseStep * factor

  const out: Array<[number, number]> = []
  for (let t = start; t <= end; t += step) {
    if (factor === 1) {
      out.push([t, values.get(t) ?? 0])
    } else {
      // Sum inside the window to preserve totals while reducing point count.
      let sum = 0
      const windowEndExclusive = Math.min(t + step, end + baseStep)
      for (let tt = t; tt < windowEndExclusive; tt += baseStep) sum += values.get(tt) ?? 0
      out.push([t, sum])
    }
  }

  if (out.length && out[out.length - 1]![0] !== end) out.push([end, values.get(end) ?? 0])
  return out
}

function buildSeriesFromBuckets(
  buckets: Array<{ bucket: number }>,
  entries: (bucket: any) => Array<[string, number]>,
  topN = 6,
  preferKeys: string[] = [],
  maxSeries = topN
): TimeSeries[] {
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

  const keys = Array.from(new Set([...preferKeys.filter(Boolean), ...topKeys])).slice(0, maxSeries)

  const timelineObserved = Array.from(new Set(buckets.map(b => b.bucket))).sort((a, b) => a - b)
  if (timelineObserved.length === 0) return []

  const inferStepSeconds = (ts: number[]) => {
    const diffs: number[] = []
    for (let i = 1; i < ts.length; i++) {
      const d = ts[i]! - ts[i - 1]!
      if (d > 0) diffs.push(d)
    }
    if (!diffs.length) return null

    // Mode of diffs is more stable than min (min is often an outlier and can explode memory).
    const freq = new Map<number, number>()
    for (const d of diffs) freq.set(d, (freq.get(d) ?? 0) + 1)
    return Array.from(freq.entries()).sort((a, b) => b[1] - a[1])[0]?.[0] ?? null
  }

  let step = inferStepSeconds(timelineObserved)
  const start = timelineObserved[0]!
  const end = timelineObserved[timelineObserved.length - 1]!

  let baseStep: number | null = step
  let factor = 1
  const timeline = (() => {
    if (!step) return timelineObserved

    const predicted = Math.floor((end - start) / step) + 1
    factor = predicted > MAX_TIMELINE_POINTS ? Math.ceil(predicted / MAX_TIMELINE_POINTS) : 1
    baseStep = step
    step = step * factor

    const out: number[] = []
    for (let t = start; t <= end; t += step) out.push(t)
    return out
  })()

  const maxSeriesByBudget = Math.max(1, Math.floor(MAX_TOTAL_POINTS / Math.max(1, timeline.length)))
  const keysLimited = keys.slice(0, Math.min(keys.length, maxSeriesByBudget))

  const series = keysLimited.map(key => ({
    name: key,
    data: timeline.map(t => {
      if (!baseStep || !step || factor === 1) {
        const val = bucketMaps.get(t)?.get(key) ?? 0
        return [t * 1000, val] as [number, number]
      }

      // Downsample window sum (preserve totals).
      let sum = 0
      const windowEndExclusive = Math.min(t + step, end + baseStep)
      for (let tt = t; tt < windowEndExclusive; tt += baseStep) sum += bucketMaps.get(tt)?.get(key) ?? 0
      return [t * 1000, sum] as [number, number]
    })
  }))

  // Hide series that are entirely zero for the selected interval.
  return series.filter(s => s.data.some(([, v]) => v > 0))
}

export function useDashboardDerived(params: {
  selectedMac: Ref<string | null>
  trafficChart: Ref<DeviceTrafficItem[]>
  domainsChart: Ref<DeviceDomainItem[] | null>
  countriesChart: Ref<DeviceCountryItem[] | null>
  protosChart: Ref<DeviceProtoItem[] | null>
  domainsTable: Ref<DeviceDomainSummary[] | null>
  countriesTable: Ref<DeviceCountrySummary[] | null>
  protosTable: Ref<DeviceProtoSummary[] | null>
}) {
  const topDomainsChart = computed<TopRow[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.domainsChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const agg = new Map<string, number>()
    for (const b of buckets) {
      for (const [domain, count] of Object.entries((b as any).domains ?? {})) {
        agg.set(domain, (agg.get(domain) ?? 0) + (Number(count ?? 0) ?? 0))
      }
    }
    return Array.from(agg.entries())
      .map(([key, value]) => ({ key, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 12)
  })

  const topCountriesChart = computed<TopRow[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.countriesChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const agg = new Map<string, number>()
    for (const b of buckets) {
      const countries = (b as any).countries
      if (Array.isArray(countries)) {
        for (const c of countries) agg.set(String(c), (agg.get(String(c)) ?? 0) + 1)
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

  const topProtosChart = computed<TopRow[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.protosChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const agg = new Map<string, number>()
    for (const b of buckets) {
      const protos = (b as any).protos
      if (Array.isArray(protos)) {
        for (const p of protos) agg.set(String(p), (agg.get(String(p)) ?? 0) + 1)
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

  const domainsTimelineSeries = computed<TimeSeries[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.domainsChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const sorted = [...buckets].sort((a, b) => a.bucket - b.bucket)
    const latest = sorted[sorted.length - 1]
    const prefer = latest
      ? Object.entries((latest as any).domains ?? {})
          .sort((a, b) => Number(b[1] ?? 0) - Number(a[1] ?? 0))
          .map(([k]) => k)
      : []
    return buildSeriesFromBuckets(
      buckets,
      b => Object.entries((b as any).domains ?? {}).map(([k, v]) => [k, Number(v ?? 0)]),
      12,
      prefer,
      12
    )
  })

  const countriesTimelineSeries = computed<TimeSeries[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.countriesChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const sorted = [...buckets].sort((a, b) => a.bucket - b.bucket)
    const latest = sorted[sorted.length - 1] as any
    const prefer =
      latest && latest.countries && typeof latest.countries === 'object' && !Array.isArray(latest.countries)
        ? Object.entries(latest.countries as Record<string, number | undefined>)
            .sort((a, b) => Number(b[1] ?? 0) - Number(a[1] ?? 0))
            .map(([k]) => k)
        : []

    const maxSeries = Math.min(30, Math.max(20, prefer.length))
    return buildSeriesFromBuckets(
      buckets,
      b => {
        const countries = (b as any).countries
        if (countries && typeof countries === 'object' && !Array.isArray(countries)) {
          return Object.entries(countries as Record<string, number | undefined>).map(([k, v]) => [k, Number(v ?? 0)])
        }
        if (Array.isArray(countries)) {
          const counts = new Map<string, number>()
          for (const c of countries) counts.set(String(c), (counts.get(String(c)) ?? 0) + 1)
          return Array.from(counts.entries())
        }
        return []
      },
      maxSeries,
      prefer,
      maxSeries
    )
  })

  const protosTimelineSeries = computed<TimeSeries[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.protosChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const sorted = [...buckets].sort((a, b) => a.bucket - b.bucket)
    const latest = sorted[sorted.length - 1] as any
    const prefer =
      latest && latest.protos && typeof latest.protos === 'object' && !Array.isArray(latest.protos)
        ? Object.entries(latest.protos as Record<string, number | undefined>)
            .sort((a, b) => Number(b[1] ?? 0) - Number(a[1] ?? 0))
            .map(([k]) => k)
        : []

    const maxSeries = Math.min(24, Math.max(16, prefer.length))
    return buildSeriesFromBuckets(
      buckets,
      b => {
        const protos = (b as any).protos
        if (protos && typeof protos === 'object' && !Array.isArray(protos)) {
          return Object.entries(protos as Record<string, number | undefined>).map(([k, v]) => [k, Number(v ?? 0)])
        }
        if (Array.isArray(protos)) {
          const counts = new Map<string, number>()
          for (const p of protos) counts.set(String(p), (counts.get(String(p)) ?? 0) + 1)
          return Array.from(counts.entries())
        }
        return []
      },
      maxSeries,
      prefer,
      maxSeries
    )
  })

  const companiesTimelineSeries = computed<TimeSeries[]>(() => {
    const mac = params.selectedMac.value
    const buckets = params.countriesChart.value?.find(x => x.device.mac === mac)?.stats ?? []
    const sorted = [...buckets].sort((a, b) => a.bucket - b.bucket)
    const latest = sorted[sorted.length - 1] as any
    const prefer =
      latest && latest.companies && typeof latest.companies === 'object' && !Array.isArray(latest.companies)
        ? Object.entries(latest.companies as Record<string, number | undefined>)
            .sort((a, b) => Number(b[1] ?? 0) - Number(a[1] ?? 0))
            .map(([k]) => k)
        : []

    const maxSeries = Math.min(30, Math.max(20, prefer.length))
    return buildSeriesFromBuckets(
      buckets,
      b => {
        const companies = (b as any).companies
        if (companies && typeof companies === 'object' && !Array.isArray(companies)) {
          return Object.entries(companies as Record<string, number | undefined>).map(([k, v]) => [k, Number(v ?? 0)])
        }
        if (Array.isArray(companies)) {
          const counts = new Map<string, number>()
          for (const c of companies) counts.set(String(c), (counts.get(String(c)) ?? 0) + 1)
          return Array.from(counts.entries())
        }
        return []
      },
      maxSeries,
      prefer,
      maxSeries
    )
  })

  const trafficSeries = computed<TimeSeries[]>(() => {
    const mac = params.selectedMac.value
    if (mac === null || mac === undefined) return []
    const item = params.trafficChart.value.find(x => x.device.mac === mac)
    if (!item) return []

    const up: Array<[number, number]> = []
    const down: Array<[number, number]> = []
    for (const b of item.stats) {
      const t = b.bucket * 1000
      up.push([t, b.up_bytes])
      down.push([t, b.down_bytes])
    }

    return [
      { name: 'Up', data: fillInternalGapsWithZero(up) },
      { name: 'Down', data: fillInternalGapsWithZero(down) }
    ]
  })

  const domainsRowsTable = computed(() => {
    const mac = params.selectedMac.value
    const stats = params.domainsTable.value?.find(x => x.device.mac === mac)?.stats
    return topFromTable(stats, 30)
  })

  const countriesRowsTable = computed(() => {
    const mac = params.selectedMac.value
    const stats = params.countriesTable.value?.find(x => x.device.mac === mac)?.stats
    return topFromTable(stats, 30)
  })

  const protosRowsTable = computed(() => {
    const mac = params.selectedMac.value
    const stats = params.protosTable.value?.find(x => x.device.mac === mac)?.stats
    return topFromTable(stats, 30)
  })

  return {
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
  }
}
