import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import type { RangeMode, RangePreset } from '@/types/range'

type PersistedRangeStateV1 = {
  v: 1
  preset: RangePreset
  rangeMode: RangeMode
  fromExpr: string
  toExpr: string
  absoluteFromMs: number | null
  absoluteToMs: number | null
}

const STORAGE_KEY = 'sniffly.range.v1'

function safeRead(): PersistedRangeStateV1 | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as PersistedRangeStateV1
    if (!parsed || parsed.v !== 1) return null
    return parsed
  } catch {
    return null
  }
}

function safeWrite(state: PersistedRangeStateV1) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch {
    // ignore (private mode / quota / disabled storage)
  }
}

export const useRangeStore = defineStore('range', () => {
  const preset = ref<RangePreset>('24h')
  const rangeMode = ref<RangeMode>('preset')

  const fromExpr = ref<string>('now-24h')
  const toExpr = ref<string>('now')

  const absoluteFromMs = ref<number | null>(null)
  const absoluteToMs = ref<number | null>(null)

  // hydrate once
  const hydrated = safeRead()
  if (hydrated) {
    preset.value = hydrated.preset
    rangeMode.value = hydrated.rangeMode
    fromExpr.value = hydrated.fromExpr
    toExpr.value = hydrated.toExpr
    absoluteFromMs.value = hydrated.absoluteFromMs
    absoluteToMs.value = hydrated.absoluteToMs
  }

  watch(
    [preset, rangeMode, fromExpr, toExpr, absoluteFromMs, absoluteToMs],
    () => {
      safeWrite({
        v: 1,
        preset: preset.value,
        rangeMode: rangeMode.value,
        fromExpr: fromExpr.value,
        toExpr: toExpr.value,
        absoluteFromMs: absoluteFromMs.value,
        absoluteToMs: absoluteToMs.value
      })
    },
    { deep: false }
  )

  return {
    preset,
    rangeMode,
    fromExpr,
    toExpr,
    absoluteFromMs,
    absoluteToMs
  }
})
